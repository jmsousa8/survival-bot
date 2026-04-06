package discord

import (
	"context"
	"testing"
	"time"

	"survival-bot/internal/config"
	"survival-bot/internal/db"
	"survival-bot/internal/games"
	"survival-bot/mocks"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// DO NOT RUN
func TestSendDailyRecap_NoDeaths(t *testing.T) {
	dbMock := &mocks.MockIDatabase{}
	dbMock.EXPECT().GetDeaths(mock.Anything, true).Return([]db.Death{}, nil)

	llmMock := &mocks.MockLlm{}

	_ = godotenv.Load("./../../.env")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	bot, err := NewBot(cfg, dbMock, llmMock)
	if err != nil {
		t.Fatalf("Failed to create new bot: %v", err)
	}

	err = bot.SendDailyRecap(context.Background())
	assert.NoError(t, err)
}

// DO NOT RUN
func TestSendDailyRecap_WithDeaths(t *testing.T) {
	dbMock := &mocks.MockIDatabase{}
	dbMock.EXPECT().GetDeaths(mock.Anything, true).Return([]db.Death{
		{Game: games.Soulmask, PlayerName: "Player1", Timestamp: time.Now()},
		{Game: games.Soulmask, PlayerName: "Player1", Timestamp: time.Now()},
		{Game: games.Soulmask, PlayerName: "Player2", Timestamp: time.Now()},
	}, nil)

	llmMock := &mocks.MockLlm{}
	llmMock.EXPECT().Ask(mock.Anything, mock.Anything).Return("llm recap message", nil)

	_ = godotenv.Load("./../../.env")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	bot, err := NewBot(cfg, dbMock, llmMock)
	if err != nil {
		t.Fatalf("Failed to create new bot: %v", err)
	}

	err = bot.SendDailyRecap(context.Background())
	assert.NoError(t, err)
}

func TestBuildLeaderboard(t *testing.T) {
	deaths := []db.Death{
		{Game: games.Soulmask, PlayerName: "Player1", Timestamp: time.Now()},
		{Game: games.Soulmask, PlayerName: "Player1", Timestamp: time.Now()},
		{Game: games.Soulmask, PlayerName: "Player2", Timestamp: time.Now()},
		{Game: games.Soulmask, PlayerName: "Player3", Timestamp: time.Now()},
	}

	cfg := &config.Config{Game: games.Soulmask}
	bot := &Bot{cfg: cfg}

	result := bot.buildLeaderboard(deaths)

	assert.Contains(t, result, "🥇 **Player1**: 2 deaths")
	assert.Contains(t, result, "🥈 **Player2**: 1 deaths")
	assert.Contains(t, result, "🥉 **Player3**: 1 deaths")
}
