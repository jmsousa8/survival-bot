package db

import (
	"fmt"
	"time"
)

type Death struct {
	ID            int64
	ServerEventID int64
	Game          string
	PlayerName    string
	KillerName    string
	DeathCause    string
	IsSuicide     bool
	Timestamp     time.Time
}

func (d *Death) String() string {
	if d.IsSuicide {
		return fmt.Sprintf(`Player "%s" died. Player killed themselves. %s`, d.PlayerName, d.DeathCause)
	}
	return fmt.Sprintf(`Player "%s" died. Was killed by "%s". %s`, d.PlayerName, d.KillerName, d.DeathCause)
}

type ServerEvent struct {
	ID        int64
	Game      string
	EventType string
	Details   string
	Timestamp time.Time
}

type PlayerLink struct {
	ID            int64
	Game          string
	PlayerName    string
	DiscordUserID string
	CreatedAt     time.Time
}
