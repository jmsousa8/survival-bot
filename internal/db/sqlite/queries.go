package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"survival-bot/internal/db"
)

func (d *database) Migrate() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS deaths (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_event_id INTEGER REFERENCES server_events(id),
			game TEXT NOT NULL,
			player_name TEXT NOT NULL,
			killer_name TEXT,
			death_cause TEXT,
			is_suicide INTEGER DEFAULT 0,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS server_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game TEXT NOT NULL,
			event_type TEXT NOT NULL,
			details TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS player_links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game TEXT NOT NULL,
			player_name TEXT NOT NULL,
			discord_user_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(game, player_name)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_deaths_player ON deaths(player_name, game);`,
		`CREATE INDEX IF NOT EXISTS idx_deaths_timestamp ON deaths(timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_deaths_server_event ON deaths(server_event_id);`,
		`CREATE INDEX IF NOT EXISTS idx_player_links_discord ON player_links(discord_user_id, game);`,
	}

	for _, schema := range schemas {
		if _, err := d.db.Exec(schema); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func (d *database) InsertDeath(death *db.Death) error {
	query := `INSERT INTO deaths (server_event_id, game, player_name, killer_name, death_cause, is_suicide, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := d.db.Exec(query, death.ServerEventID, death.Game, death.PlayerName, death.KillerName, death.DeathCause, death.IsSuicide, death.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert death: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	death.ID = id
	return nil
}

func (d *database) GetDeaths(game string, today bool) ([]db.Death, error) {
	query := fmt.Sprintf(`
		SELECT d.id, d.server_event_id, d.game, d.player_name, d.killer_name, d.death_cause, d.is_suicide, d.timestamp
		FROM deaths d
		WHERE d.game = ? 
		%s
		ORDER BY d.timestamp DESC`, d.buildTodayFilter(today))

	rows, err := d.db.Query(query, game)
	if err != nil {
		return nil, fmt.Errorf("failed to query deaths: %w", err)
	}
	defer rows.Close()

	var deaths []db.Death
	for rows.Next() {
		var death db.Death
		err = rows.Scan(&death.ID, &death.ServerEventID, &death.Game, &death.PlayerName, &death.KillerName, &death.DeathCause, &death.IsSuicide, &death.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan death: %w", err)
		}
		deaths = append(deaths, death)
	}

	return deaths, nil
}

func (d *database) InsertServerEvent(event *db.ServerEvent) error {
	query := `INSERT INTO server_events (game, event_type, details, timestamp) VALUES (?, ?, ?, ?)`
	result, err := d.db.Exec(query, event.Game, event.EventType, event.Details, event.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert server event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	event.ID = id
	return nil
}

func (d *database) InsertPlayerLink(link *db.PlayerLink) error {
	query := `INSERT OR REPLACE INTO player_links (game, player_name, discord_user_id, created_at) VALUES (?, ?, ?, ?)`
	result, err := d.db.Exec(query, link.Game, link.PlayerName, link.DiscordUserID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert player link: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	link.ID = id
	return nil
}

func (d *database) DeletePlayerLink(game, playerName string) error {
	query := `DELETE FROM player_links WHERE game = ? AND player_name = ?`
	_, err := d.db.Exec(query, game, playerName)
	if err != nil {
		return fmt.Errorf("failed to delete player link: %w", err)
	}
	return nil
}

func (d *database) GetPlayerLinksByGame(game string) ([]db.PlayerLink, error) {
	query := `SELECT id, game, player_name, discord_user_id, created_at FROM player_links WHERE game = ?`

	rows, err := d.db.Query(query, game)
	if err != nil {
		return nil, fmt.Errorf("failed to query player links: %w", err)
	}
	defer rows.Close()

	var links []db.PlayerLink
	for rows.Next() {
		var link db.PlayerLink
		err := rows.Scan(&link.ID, &link.Game, &link.PlayerName, &link.DiscordUserID, &link.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player link: %w", err)
		}
		links = append(links, link)
	}

	return links, nil
}

func (d *database) GetPlayerLinkByDiscord(discordUserID, game string) (*db.PlayerLink, error) {
	query := `SELECT id, game, player_name, discord_user_id, created_at FROM player_links WHERE discord_user_id = ? AND game = ?`

	var link db.PlayerLink
	err := d.db.QueryRow(query, discordUserID, game).Scan(&link.ID, &link.Game, &link.PlayerName, &link.DiscordUserID, &link.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player link: %w", err)
	}
	return &link, nil
}

func (d *database) buildTodayFilter(today bool) string {
	if today {
		return "AND date(d.timestamp) = date('now')"
	}
	return ""
}

func (d *database) CountDeathsByFallDamage(game string, today bool) (*db.CauseOfDeathStats, error) {
	query := fmt.Sprintf(`
		SELECT d.player_name, COUNT(*) as count
		FROM deaths d
		WHERE d.game = ? AND d.death_cause LIKE '%%fall damage%%'
		%s
		GROUP BY d.player_name
		ORDER BY count DESC`, d.buildTodayFilter(today))

	rows, err := d.db.Query(query, game)
	if err != nil {
		return nil, fmt.Errorf("failed to query fall deaths: %w", err)
	}
	defer rows.Close()

	stats := &db.CauseOfDeathStats{
		Total:             0,
		DeathCause:        "fall damage",
		PlayerNameCounter: make(map[string]int),
	}
	var total int
	for rows.Next() {
		var player string
		var cnt int
		err = rows.Scan(&player, &cnt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fall death: %w", err)
		}
		stats.PlayerNameCounter[player] = cnt
		total += cnt
	}
	stats.Total = total

	return stats, nil
}

func (d *database) CountDeathsBySuicide(game string, today bool) (*db.CauseOfDeathStats, error) {
	query := fmt.Sprintf(`
		SELECT d.player_name, COUNT(*) as count
		FROM deaths d
		WHERE d.game = ? AND d.is_suicide = 1
		%s
		GROUP BY d.player_name
		ORDER BY count DESC`, d.buildTodayFilter(today))

	rows, err := d.db.Query(query, game)
	if err != nil {
		return nil, fmt.Errorf("failed to query suicides: %w", err)
	}
	defer rows.Close()

	stats := &db.CauseOfDeathStats{
		Total:             0,
		DeathCause:        "suicide",
		PlayerNameCounter: make(map[string]int),
	}
	var total int
	for rows.Next() {
		var player string
		var cnt int
		err = rows.Scan(&player, &cnt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan suicide: %w", err)
		}
		stats.PlayerNameCounter[player] = cnt
		total += cnt
	}
	stats.Total = total

	return stats, nil
}

func (d *database) GetMostCommonCauseOfDeath(game string, today bool) (*db.CauseOfDeathStats, error) {
	query := fmt.Sprintf(`
		SELECT d.death_cause, COUNT(*) as count
		FROM deaths d
		WHERE d.game = ? 
		%s
		GROUP BY d.death_cause
		ORDER BY count DESC LIMIT 1`, d.buildTodayFilter(today))

	var deathCause string
	var count int
	err := d.db.QueryRow(query, game).Scan(&deathCause, &count)
	if err != nil {
		return nil, fmt.Errorf("failed to query common cause of death: %w", err)
	}

	query = fmt.Sprintf(`
		SELECT d.player_name, COUNT(*) as count
		FROM deaths d
		WHERE d.game = ? AND d.death_cause = ? 
		%s
		GROUP BY d.player_name
		ORDER BY count DESC`, d.buildTodayFilter(today))

	rows, err := d.db.Query(query, game, deathCause)
	if err != nil {
		return nil, fmt.Errorf("failed to query deaths by cause: %w", err)
	}
	defer rows.Close()

	stats := &db.CauseOfDeathStats{
		Total:             0,
		DeathCause:        deathCause,
		PlayerNameCounter: make(map[string]int),
	}
	var total int
	for rows.Next() {
		var player string
		var cnt int
		err = rows.Scan(&player, &cnt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan death by cause: %w", err)
		}
		stats.PlayerNameCounter[player] = cnt
		total += cnt
	}
	stats.Total = total

	return stats, nil
}

func (d *database) GetMostCommonKiller(game string, today bool) (*db.KillerStats, error) {
	query := fmt.Sprintf(`
		SELECT d.killer_name, COUNT(*) as count
		FROM deaths d
		WHERE d.game = ? AND d.is_suicide = 0 
		%s
		GROUP BY d.killer_name
		ORDER BY count DESC LIMIT 1`, d.buildTodayFilter(today))

	var killer string
	var count int
	err := d.db.QueryRow(query, game).Scan(&killer, &count)
	if err != nil {
		return nil, fmt.Errorf("failed to query common killer: %w", err)
	}

	query = fmt.Sprintf(`
		SELECT d.player_name, COUNT(*) as count
		FROM deaths d
		WHERE d.game = ? AND d.killer_name = ? 
		%s
		GROUP BY d.player_name
		ORDER BY count DESC`, d.buildTodayFilter(today))

	rows, err := d.db.Query(query, game, killer)
	if err != nil {
		return nil, fmt.Errorf("failed to query deaths by killer: %w", err)
	}
	defer rows.Close()

	stats := &db.KillerStats{
		KillerName:        killer,
		Total:             0,
		PlayerNameCounter: make(map[string]int),
	}
	var total int
	for rows.Next() {
		var player string
		var cnt int
		err = rows.Scan(&player, &cnt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan death by killer: %w", err)
		}
		stats.PlayerNameCounter[player] = cnt
		total += cnt
	}
	stats.Total = total

	return stats, nil
}

func (d *database) GetPlayerStats(game, playerName string) (*db.PlayerStats, error) {
	var stats db.PlayerStats
	stats.PlayerName = playerName

	totalQuery := `SELECT COUNT(*) FROM deaths WHERE player_name = ? AND game = ?`
	err := d.db.QueryRow(totalQuery, playerName, game).Scan(&stats.TotalDeaths)
	if err != nil {
		return nil, fmt.Errorf("failed to get total deaths: %w", err)
	}

	rankQuery := `
		SELECT COUNT(*) + 1 FROM (
			SELECT player_name, COUNT(*) as cnt
			FROM deaths
			WHERE game = ?
			GROUP BY player_name
			HAVING cnt > (SELECT COUNT(*) FROM deaths WHERE player_name = ? AND game = ?)
		)`
	err = d.db.QueryRow(rankQuery, game, playerName, game).Scan(&stats.Rank)
	if err != nil {
		return nil, fmt.Errorf("failed to get rank: %w", err)
	}

	fallQuery := `SELECT COUNT(*) FROM deaths WHERE player_name = ? AND game = ? AND death_cause LIKE '%fall%'`
	err = d.db.QueryRow(fallQuery, playerName, game).Scan(&stats.FallDeaths)
	if err != nil {
		return nil, fmt.Errorf("failed to get fall deaths: %w", err)
	}

	suicideQuery := `SELECT COUNT(*) FROM deaths WHERE player_name = ? AND game = ? AND is_suicide = 1`
	err = d.db.QueryRow(suicideQuery, playerName, game).Scan(&stats.Suicides)
	if err != nil {
		return nil, fmt.Errorf("failed to get suicides: %w", err)
	}

	causeQuery := `
		SELECT death_cause, COUNT(*) as cnt
		FROM deaths
		WHERE player_name = ? AND game = ? AND death_cause IS NOT NULL AND death_cause != ''
		GROUP BY death_cause
		ORDER BY cnt DESC
		LIMIT 1`
	var causeCount int
	err = d.db.QueryRow(causeQuery, playerName, game).Scan(&stats.MostCommonCause, &causeCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get most common cause: %w", err)
	}

	enemyQuery := `
		SELECT killer_name, COUNT(*) as cnt
		FROM deaths
		WHERE player_name = ? AND game = ? AND killer_name IS NOT NULL AND killer_name != '' AND is_suicide = 0
		GROUP BY killer_name
		ORDER BY cnt DESC
		LIMIT 1`
	var enemyCount int
	err = d.db.QueryRow(enemyQuery, playerName, game).Scan(&stats.DeadliestEnemy, &enemyCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get deadliest enemy: %w", err)
	}

	return &stats, nil
}

func (d *database) GetLastDeath(game, playerName string) (*db.Death, error) {
	query := `SELECT id, server_event_id, game, player_name, killer_name, death_cause, is_suicide, timestamp 
		FROM deaths 
		WHERE player_name = ? AND game = ? 
		ORDER BY timestamp DESC 
		LIMIT 1`

	var death db.Death
	err := d.db.QueryRow(query, playerName, game).Scan(
		&death.ID, &death.ServerEventID, &death.Game, &death.PlayerName,
		&death.KillerName, &death.DeathCause, &death.IsSuicide, &death.Timestamp,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last death: %w", err)
	}

	return &death, nil
}

func (d *database) GetLastDeaths(game, playerName string, limit int64) ([]db.Death, error) {
	query := `SELECT id, server_event_id, game, player_name, killer_name, death_cause, is_suicide, timestamp
		FROM deaths 
		WHERE game = ? AND player_name = ? 
		ORDER BY timestamp DESC
		LIMIT ?`

	rows, err := d.db.Query(query, game, playerName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query last deaths: %w", err)
	}
	defer rows.Close()

	var deaths []db.Death
	for rows.Next() {
		var death db.Death
		err = rows.Scan(&death.ID, &death.ServerEventID, &death.Game, &death.PlayerName, &death.KillerName, &death.DeathCause, &death.IsSuicide, &death.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan death: %w", err)
		}
		deaths = append(deaths, death)
	}

	return deaths, nil
}
