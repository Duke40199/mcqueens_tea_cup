package database

import (
	"McQueens_Tea_Cup/internal/config"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Database interface{}

type PostgresDB struct {
}

func NewPostgresDBConn(cfg config.DatabaseConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	dbConn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	dbConn.SetMaxOpenConns(10) // adjust based on Supabase limit
	dbConn.SetMaxIdleConns(5)
	dbConn.SetConnMaxLifetime(time.Hour)
	return dbConn, nil
}
