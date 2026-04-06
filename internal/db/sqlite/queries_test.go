package sqlite

import (
	"testing"
	"time"

	"survival-bot/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *database {
	db, err := New(":memory:")
	require.NoError(t, err)

	d, ok := db.(*database)
	require.True(t, ok)

	err = d.Migrate()
	require.NoError(t, err)

	return d
}

func TestInsertDeath(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	death := &db.Death{
		Game:       "testgame",
		PlayerName: "Player1",
		KillerName: "Enemy1",
		DeathCause: "fall damage",
		IsSuicide:  false,
	}

	err := d.InsertDeath(death)
	require.NoError(t, err)
	assert.Greater(t, death.ID, int64(0))
}

func TestGetDeaths_AllTime(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Enemy1", DeathCause: "fall damage", IsSuicide: false})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Enemy2", DeathCause: "melee", IsSuicide: false})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player2", KillerName: "Player2", IsSuicide: true})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "othergame", PlayerName: "Player3", KillerName: "Enemy1"})
	require.NoError(t, err)

	deaths, err := d.GetDeaths("testgame", false)
	require.NoError(t, err)
	assert.Len(t, deaths, 3)
}

func TestGetDeaths_Today(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Enemy1", DeathCause: "fall damage", Timestamp: time.Now().AddDate(0, 0, -1)})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Enemy2", Timestamp: time.Now()})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player2", KillerName: "Enemy2", Timestamp: time.Now()})
	require.NoError(t, err)

	deaths, err := d.GetDeaths("testgame", true)
	require.NoError(t, err)
	assert.Len(t, deaths, 2)
}

func TestCountDeathsByFallDamage(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", DeathCause: "Died of fall damage"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", DeathCause: "Died of fall damage"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player2", DeathCause: "Died of fall damage"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player3", KillerName: "Enemy1", DeathCause: "melee"})
	require.NoError(t, err)

	stats, err := d.CountDeathsByFallDamage("testgame", false)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.Total)

	assert.Equal(t, 2, stats.PlayerNameCounter["Player1"])
	assert.Equal(t, 1, stats.PlayerNameCounter["Player2"])
}

func TestCountDeathsBySuicide(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", IsSuicide: true})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", IsSuicide: true})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player2", IsSuicide: true})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player3", KillerName: "Enemy1", IsSuicide: false})
	require.NoError(t, err)

	stats, err := d.CountDeathsBySuicide("testgame", false)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.Total)

	assert.Equal(t, 2, stats.PlayerNameCounter["Player1"])
	assert.Equal(t, 1, stats.PlayerNameCounter["Player2"])
}

func TestGetMostCommonCauseOfDeath(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", DeathCause: "melee"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", DeathCause: "melee"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player2", DeathCause: "melee"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player3", DeathCause: "fall damage"})
	require.NoError(t, err)

	stats, err := d.GetMostCommonCauseOfDeath("testgame", false)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, "melee", stats.DeathCause)

	assert.Equal(t, 2, stats.PlayerNameCounter["Player1"])
	assert.Equal(t, 1, stats.PlayerNameCounter["Player2"])
}

func TestGetMostCommonKiller(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Wolf"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Wolf"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player2", KillerName: "Wolf"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player3", KillerName: "Troll"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player4", IsSuicide: true})
	require.NoError(t, err)

	stats, err := d.GetMostCommonKiller("testgame", false)
	require.NoError(t, err)
	assert.Equal(t, "Wolf", stats.KillerName)
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 2, stats.PlayerNameCounter["Player1"])
	assert.Equal(t, 1, stats.PlayerNameCounter["Player2"])
}

func TestGetPlayerStats(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", DeathCause: "fall damage", IsSuicide: false})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", DeathCause: "fall damage", IsSuicide: false})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Player1", IsSuicide: true})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", KillerName: "Goblin"})
	require.NoError(t, err)
	err = d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player2", KillerName: "Goblin"})
	require.NoError(t, err)

	stats, err := d.GetPlayerStats("testgame", "Player1")
	require.NoError(t, err)

	assert.Equal(t, "Player1", stats.PlayerName)
	assert.Equal(t, 4, stats.TotalDeaths)
	assert.Equal(t, 1, stats.Rank)
	assert.Equal(t, 2, stats.FallDeaths)
	assert.Equal(t, 1, stats.Suicides)
	assert.Equal(t, "fall damage", stats.MostCommonCause)
	assert.Equal(t, "Goblin", stats.DeadliestEnemy)
}

func TestInsertServerEvent(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	event := &db.ServerEvent{
		Game:      "testgame",
		EventType: "PLAYER_DIED",
		Details:   `{"player": "test"}`,
	}

	err := d.InsertServerEvent(event)
	require.NoError(t, err)
	assert.Greater(t, event.ID, int64(0))
}

func TestPlayerLinks(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	link := &db.PlayerLink{
		Game:          "testgame",
		PlayerName:    "Player1",
		DiscordUserID: "123456",
	}

	err := d.InsertPlayerLink(link)
	require.NoError(t, err)
	assert.Greater(t, link.ID, int64(0))

	retrieved, err := d.GetPlayerLinkByDiscord("123456", "testgame")
	require.NoError(t, err)
	assert.Equal(t, "Player1", retrieved.PlayerName)

	links, err := d.GetPlayerLinksByGame("testgame")
	require.NoError(t, err)
	assert.Len(t, links, 1)

	err = d.DeletePlayerLink("testgame", "Player1")
	require.NoError(t, err)

	retrieved, err = d.GetPlayerLinkByDiscord("123456", "testgame")
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestGetDeaths_Empty(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	deaths, err := d.GetDeaths("testgame", false)
	require.NoError(t, err)
	assert.Len(t, deaths, 0)
}

func TestGetPlayerStats_NoData(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	stats, err := d.GetPlayerStats("testgame", "Nonexistent")
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalDeaths)
	assert.Equal(t, 1, stats.Rank)
}

func TestGetMostCommonCauseOfDeath_NoData(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	stats, err := d.GetMostCommonCauseOfDeath("testgame", false)
	require.Error(t, err)
	assert.Nil(t, stats)
}

func TestGetMostCommonKiller_NoSuicides(t *testing.T) {
	d := setupTestDB(t)
	defer d.Close()

	err := d.InsertDeath(&db.Death{Game: "testgame", PlayerName: "Player1", IsSuicide: true})
	require.NoError(t, err)

	stats, err := d.GetMostCommonKiller("testgame", false)
	require.Error(t, err)
	assert.Nil(t, stats)
}
