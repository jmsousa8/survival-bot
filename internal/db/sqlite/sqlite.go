package sqlite

import (
	"database/sql"

	"survival-bot/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

type database struct {
	db *sql.DB
}

func New(dbPath string) (db.IDatabase, error) {
	dbsql, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err = dbsql.Ping(); err != nil {
		return nil, err
	}

	return &database{db: dbsql}, nil
}

func (d *database) Close() error {
	return d.db.Close()
}
