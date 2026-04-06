package valheim

import (
	"fmt"
	"math/rand/v2"
	"time"

	"survival-bot/internal/events"

	"github.com/bwmarrin/discordgo"
)

type playerDiedEvent struct {
	Timestamp  time.Time
	PlayerName string
	RawSource  string
	Message    string
}

func newPlayerDiedEvent(raw, playerName string) *playerDiedEvent {
	return &playerDiedEvent{
		Timestamp:  time.Now(),
		PlayerName: playerName,
		RawSource:  raw,
		Message:    generateMessage(playerName),
	}
}

func (e *playerDiedEvent) Type() events.EventType {
	return events.EventPlayerDied
}

func (e *playerDiedEvent) Raw() string {
	return e.RawSource
}

func (e *playerDiedEvent) GetMessage() string {
	return e.Message
}

func (e *playerDiedEvent) SetMessage(message string) {
	e.Message = message
}

func (e *playerDiedEvent) GetPlayer() string {
	return e.PlayerName
}

func (e *playerDiedEvent) GetKiller() string {
	return ""
}

func (e *playerDiedEvent) GetDeathCause() string {
	return ""
}

func (e *playerDiedEvent) ToDiscordEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Player Died",
		Description: e.GetMessage(),
		Color:       0xFF0000,
	}
}

func (e *playerDiedEvent) Prompt() string {
	return fmt.Sprintf(`Generate a short, dramatic message about a death that occurred in game.

Requirements:
- 1-2 sentences max
- Game: Valheim (https://valheim.fandom.com/wiki/Valheim_(world))
- Player: %s
- ROAST THEM HARD for being so bad at the game!
- Include 1 relevant emoji
- Format: just the message, no quotes or extra text, player name should be wrapped in ** **

Just output the message, nothing else.`, e.PlayerName)
}

func generateMessage(playerName string) string {
	var deathMessages = []string{
		fmt.Sprintf("⚔ **{%s}** has fallen in battle!", playerName),
		fmt.Sprintf("🪦 **{%s}** is now food for worms...", playerName),
		fmt.Sprintf("💀 **{%s}** bit the dust...", playerName),
		fmt.Sprintf("🏹 **{%s}** traveled to Valhalla... prematurely", playerName),
	}

	return deathMessages[rand.IntN(len(deathMessages))]
}
