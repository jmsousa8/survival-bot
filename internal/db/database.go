package db

type CauseOfDeathStats struct {
	DeathCause        string
	Total             int
	PlayerNameCounter map[string]int
}

type KillerStats struct {
	KillerName        string
	Total             int
	PlayerNameCounter map[string]int
}

type PlayerStats struct {
	PlayerName      string
	TotalDeaths     int
	Rank            int
	FallDeaths      int
	Suicides        int
	MostCommonCause string
	DeadliestEnemy  string
}

type IDatabase interface {
	Close() error
	Migrate() error

	InsertDeath(death *Death) error
	GetDeaths(game string, today bool) ([]Death, error)
	GetLastDeath(game, playerName string) (*Death, error)

	CountDeathsByFallDamage(game string, today bool) (*CauseOfDeathStats, error)
	CountDeathsBySuicide(game string, today bool) (*CauseOfDeathStats, error)
	GetMostCommonCauseOfDeath(game string, today bool) (*CauseOfDeathStats, error)

	GetMostCommonKiller(game string, today bool) (*KillerStats, error)
	GetPlayerStats(game, playerName string) (*PlayerStats, error)

	InsertServerEvent(event *ServerEvent) error

	InsertPlayerLink(link *PlayerLink) error
	DeletePlayerLink(game, playerName string) error
	GetPlayerLinksByGame(game string) ([]PlayerLink, error)
	GetPlayerLinkByDiscord(discordUserID, game string) (*PlayerLink, error)
}
