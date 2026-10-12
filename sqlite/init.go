package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func apply(db *sql.DB, file string, n int) error {
	query, err := migrationFS.ReadFile(path.Join("migrations", file))
	if err != nil {
		return fmt.Errorf("read migration %s! %w", file, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", file, err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec(string(query)); err != nil {
		return fmt.Errorf("run migration %s! %w", file, err)
	}

	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", n)); err != nil {
		return fmt.Errorf("set user version after %s! %w", file, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s! %w", file, err)
	}

	return nil
}

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	if version > len(entries) {
		return fmt.Errorf("database is at version %d, but %d migrations are known", version, len(entries))
	}

	for i, e := range entries {
		if i < version {
			continue
		}

		n := i + 1

		if err := apply(db, e.Name(), n); err != nil {
			return err
		}
	}

	return nil
}

func InitDb() *sql.DB {
	db, err := sql.Open("sqlite3", "./subscribehn.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}

	return db
}
