package discord

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"survival-bot/internal/ai"
	"survival-bot/internal/config"
	"survival-bot/internal/db"
	"survival-bot/internal/events"

	"github.com/bwmarrin/discordgo"
)

type IDiscord interface {
	SendMessage(event events.Event) error
	SendDailyRecap(ctx context.Context) error
	Close() error
	Game() string
	AddDebug(msg string)
}

type Bot struct {
	session       *discordgo.Session
	cfg           *config.Config
	db            db.IDatabase
	llm           ai.Llm
	handlers      map[string]CommandHandlerWithResp
	debugMessages []string
	mu            sync.Mutex
}

type CommandHandlerWithResp func(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse

func NewBot(cfg *config.Config, database db.IDatabase, llm ai.Llm) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Discord bot logged in as %s\n", r.User.Username)
	})

	bot := &Bot{
		session:  session,
		cfg:      cfg,
		db:       database,
		llm:      llm,
		handlers: make(map[string]CommandHandlerWithResp),
	}

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		bot.handleDM(s, m)
	})

	return bot, nil
}

func (b *Bot) RegisterCommand(name string, handler CommandHandlerWithResp) {
	b.handlers[name] = handler
}

func (b *Bot) SetupCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "stats",
			Description: "Show player death stats",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "player",
					Description: "Discord user to check stats for",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		{
			Name:        "leaderboard",
			Description: "Show deaths leaderboard",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "timeframe",
					Description: "Time frame for the leaderboard",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Today", Value: "today"},
						{Name: "All Time", Value: "alltime"},
					},
				},
			},
		},
		{
			Name:        "roast",
			Description: "Roast a player based on their stats",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "player",
					Description: "Discord user to roast",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		{
			Name:        "deaths",
			Description: "Show player last deaths",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "player",
					Description: "Discord user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
				{
					Name:        "count",
					Description: "How many deaths to list. Defaults to 10",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
			},
		},
	}

	for _, cmd := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.cfg.DiscordGuildID, cmd)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmd.Name, err)
		}
		log.Printf("Registered command: /%s\n", cmd.Name)
	}

	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		handler, ok := b.handlers[i.ApplicationCommandData().Name]
		if !ok {
			return
		}

		response := handler(s, i)
		if response != nil {
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				log.Printf("Failed to respond to interaction: %v\n", err)
			}
		}
	})

	return nil
}

func (b *Bot) Open() error {
	return b.session.Open()
}

func (b *Bot) Close() error {
	return b.session.Close()
}

func (b *Bot) SendMessage(event events.Event) error {
	_, err := b.session.ChannelMessageSendEmbed(b.cfg.DiscordChannelID, event.ToDiscordEmbed())
	return err
}

func (b *Bot) Game() string {
	return b.cfg.Game
}

func (b *Bot) AddDebug(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	b.debugMessages = append(b.debugMessages, fmt.Sprintf("[%s] %s", timestamp, msg))
}
