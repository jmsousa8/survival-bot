package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"survival-bot/internal/ai"
	"survival-bot/internal/config"
	"survival-bot/internal/db"
	"survival-bot/internal/db/sqlite"
	"survival-bot/internal/discord"
	"survival-bot/internal/events"
	_ "survival-bot/internal/games/soulmask"
	_ "survival-bot/internal/games/valheim"
	"survival-bot/internal/logwatcher"
)

type App struct {
	cfg        *config.Config
	db         db.IDatabase
	discordBot discord.IDiscord
	logWatcher logwatcher.IWatcher
	llm        ai.Llm
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	database, err := sqlite.New(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	if err = database.Migrate(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	var gemini ai.Llm
	if cfg.GeminiAPIKey != "" {
		g, err := ai.NewGemini(ctx, cfg.GeminiAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create gemini client: %w", err)
		}
		gemini = g
	}

	discordBot, err := discord.NewBot(cfg, database, gemini)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord bot: %w", err)
	}

	discordBot.RegisterCommand("stats", discordBot.HandleStats)
	discordBot.RegisterCommand("leaderboard", discordBot.HandleLeaderboard)
	discordBot.RegisterCommand("roast", discordBot.HandleRoast)
	discordBot.RegisterCommand("deaths", discordBot.HandleDeaths)

	if err = discordBot.Open(); err != nil {
		return nil, fmt.Errorf("failed to open discord connection: %w", err)
	}

	if err = discordBot.SetupCommands(); err != nil {
		_ = discordBot.Close()
		return nil, fmt.Errorf("failed to setup commands: %w", err)
	}

	logWatcher, err := logwatcher.NewWatcher(cfg.LogFilePath, cfg.Game)
	if err != nil {
		_ = discordBot.Close()
		return nil, fmt.Errorf("failed to create log watcher: %w", err)
	}

	return &App{
		cfg:        cfg,
		db:         database,
		discordBot: discordBot,
		logWatcher: logWatcher,
		llm:        gemini,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log.Printf("Game: %s\n", a.cfg.Game)
	log.Printf("Watching log file: %s\n", a.cfg.LogFilePath)
	log.Println("Bot is ready. Press Ctrl+C to stop.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastRecapDate := time.Now().Day()

	for {
		select {
		case <-sigChan:
			log.Println("Shutting down...")
			return nil

		case <-ticker.C:
			a.processLogEvents(ctx)

			now := time.Now()
			if now.Day() != lastRecapDate && now.Hour() == 0 && now.Minute() == 0 {
				log.Println("Sending daily recap...")
				if err := a.discordBot.SendDailyRecap(ctx); err != nil {
					a.discordBot.AddDebug(fmt.Sprintf("DailyRecap: %v", err))
					log.Printf("Error sending daily recap: %v\n", err)
				}
				lastRecapDate = now.Day()
			}
		}
	}
}

func (a *App) Close() error {
	if a.db != nil {
		_ = a.db.Close()
	}
	if a.logWatcher != nil {
		_ = a.logWatcher.Close()
	}
	if a.discordBot != nil {
		return a.discordBot.Close()
	}
	return nil
}

func (a *App) processLogEvents(ctx context.Context) {
	rawEvents, err := a.logWatcher.ReadNewLines()
	if err != nil {
		a.discordBot.AddDebug(fmt.Sprintf("ReadNewLines: %v", err))
		log.Printf("Error reading log: %v\n", err)
		return
	}

	for _, event := range rawEvents {
		log.Printf("Event detected: %s\n", event.Type())

		if a.llm != nil && event.Prompt() != "" {
			genMsg, err := a.llm.Ask(ctx, event.Prompt())
			if err != nil {
				a.discordBot.AddDebug(fmt.Sprintf("llm.Ask: %v", err))
				log.Printf("Error asking gemini: %v\n", err)
			} else {
				event.SetMessage(genMsg)
			}
		}

		if err = a.saveServerEvent(event); err != nil {
			a.discordBot.AddDebug(fmt.Sprintf("saveServerEvent: %v", err))
			log.Printf("Error saving event: %v\n", err)
		}

		if err = a.discordBot.SendMessage(event); err != nil {
			a.discordBot.AddDebug(fmt.Sprintf("SendMessage: %v", err))
			log.Printf("Error sending event: %v\n", err)
		}
	}
}

func (a *App) saveServerEvent(event events.Event) error {
	details, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	serverEvent := &db.ServerEvent{
		Game:      a.cfg.Game,
		EventType: string(event.Type()),
		Details:   string(details),
		Timestamp: time.Now(),
	}
	if err = a.db.InsertServerEvent(serverEvent); err != nil {
		return fmt.Errorf("failed to insert server event: %w", err)
	}

	if d := event.HasDebug(); d != "" {
		a.discordBot.AddDebug(d)
	}

	if deathEvent, ok := event.(events.DeathEvent); ok {
		death := &db.Death{
			ServerEventID: serverEvent.ID,
			Game:          a.cfg.Game,
			PlayerName:    deathEvent.GetPlayer(),
			KillerName:    deathEvent.GetKiller(),
			DeathCause:    deathEvent.GetDeathCause(),
			IsSuicide:     deathEvent.GetPlayer() == deathEvent.GetKiller(),
			Timestamp:     time.Now(),
		}
		if err = a.db.InsertDeath(death); err != nil {
			return fmt.Errorf("failed to insert death: %w", err)
		}
	}

	return nil
}
