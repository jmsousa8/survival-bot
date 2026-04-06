package config

import (
	"errors"
	"os"

	"survival-bot/internal/games"

	"github.com/joho/godotenv"
)

type Config struct {
	Game             string
	LogFilePath      string
	DiscordBotToken  string
	DiscordGuildID   string
	DiscordChannelID string
	DBPath           string
	GeminiAPIKey     string
	OwnerID          string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	game := os.Getenv("GAME")
	if !games.IsValid(game) {
		return nil, errors.New("GAME not valid")
	}

	logPath := os.Getenv("LOG_FILE_PATH")
	if logPath == "" {
		return nil, errors.New("LOG_FILE_PATH not set")
	}

	discordBotToken := os.Getenv("DISCORD_BOT_TOKEN")
	if discordBotToken == "" {
		return nil, errors.New("DISCORD_BOT_TOKEN not set")
	}

	discordGuildID := os.Getenv("DISCORD_GUILD_ID")
	if discordGuildID == "" {
		return nil, errors.New("DISCORD_GUILD_ID not set")
	}

	discordChannelID := os.Getenv("DISCORD_CHANNEL_ID")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./bot.db"
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	ownerID := os.Getenv("DISCORD_OWNER_ID")

	return &Config{
		Game:             game,
		LogFilePath:      logPath,
		DiscordBotToken:  discordBotToken,
		DiscordGuildID:   discordGuildID,
		DiscordChannelID: discordChannelID,
		DBPath:           dbPath,
		GeminiAPIKey:     geminiAPIKey,
		OwnerID:          ownerID,
	}, nil
}
