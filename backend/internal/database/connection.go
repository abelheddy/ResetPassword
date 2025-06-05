package database

import (
	"database/sql"
	"fmt"
	"time"
	"password-recovery/internal/config"

	_ "github.com/lib/pq"
)

func InitDB(cfg config.AppConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error al conectar a la base de datos: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al verificar conexión a la BD: %v", err)
	}

	return db, nil
}