package database

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS smtp_config (
			id SERIAL PRIMARY KEY,
			host VARCHAR(100) NOT NULL,
			port INTEGER NOT NULL,
			username VARCHAR(100) NOT NULL,
			password VARCHAR(100) NOT NULL,
			from_email VARCHAR(100) NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_smtp_config_active ON smtp_config(is_active)`,
		`CREATE TABLE IF NOT EXISTS reset_codes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			code VARCHAR(10) NOT NULL,
			expiration_time TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_reset_codes_expiration ON reset_codes(expiration_time)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("error al ejecutar migración: %v\nSQL: %s", err, migration)
		}
	}

	return nil
}