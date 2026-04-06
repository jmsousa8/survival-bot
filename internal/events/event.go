package events

import "github.com/bwmarrin/discordgo"

type EventType string

const (
	EventPlayerDied  EventType = "PLAYER_DIED"
	EventServerStart EventType = "SERVER_START"
	EventServerStop  EventType = "SERVER_STOP"
	EventBossKill    EventType = "BOSS_KILL"
)

type Event interface {
	Type() EventType
	Raw() string
	GetMessage() string
	SetMessage(message string)
	Prompt() string
	ToDiscordEmbed() *discordgo.MessageEmbed
	HasDebug() string
}

type DeathEvent interface {
	GetPlayer() string
	GetKiller() string
	GetDeathCause() string
}
