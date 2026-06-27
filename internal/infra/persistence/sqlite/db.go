package sqlite

import (
	"database/sql"
	"embed"

	_ "github.com/mattn/go-sqlite3"
)

var migrationsFS embed.FS

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	migrations, err := migrationsFS.ReadDir(".")
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		if !migration.IsDir() {
			content, err := migrationsFS.ReadFile(migration.Name())
			if err != nil {
				return err
			}
			_, err = db.Exec(string(content))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
