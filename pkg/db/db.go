package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT '',
	title VARCHAR(255) NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	repeat VARCHAR(128) NOT NULL DEFAULT '' CHECK (length(repeat) <= 128)
);
CREATE INDEX IF NOT EXISTS scheduler_date_idx ON scheduler (date);
`

var db *sql.DB

func Init(dbFile string) (err error) {
	if db != nil {
		return errors.New("database is already initialized")
	}
	info, err := os.Stat(dbFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check database file: %w", err)
	}
	if err == nil && info.IsDir() {
		return fmt.Errorf("database path is a directory: %s", dbFile)
	}

	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, conn.Close())
		}
	}()

	if err = conn.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	if _, err = conn.Exec(schema); err != nil {
		return fmt.Errorf("initialize database schema: %w", err)
	}
	db = conn
	return nil
}

func Close() error {
	if db == nil {
		return nil
	}
	err := db.Close()
	db = nil
	return err
}
