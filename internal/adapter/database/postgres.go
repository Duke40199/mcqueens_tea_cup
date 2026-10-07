package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"McQueens_Tea_Cup/internal/config"
)

type Database interface{}

type PostgresDB struct {
}

func NewPostgresDBConn(cfg config.DatabaseConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	dbConn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		// Return the error so the caller decides how to handle it, instead of
		// killing the process from inside a constructor.
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}
	dbConn.SetMaxOpenConns(10)
	dbConn.SetMaxIdleConns(5)
	dbConn.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbConn.PingContext(ctx); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return dbConn, nil
}
