package db

import (
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var ErrNotFound = fmt.Errorf("resource not found")

// Config represents DB configs
type Config struct {
	Host     string `json:"host" envconfig:"DB_HOST" default:"localhost"`
	Port     int    `json:"port" envconfig:"DB_PORT" default:"5432"`
	User     string `json:"user" envconfig:"DB_USER" default:"postgres"`
	Password string `json:"password" envconfig:"DB_PASSWORD" default:"postgres"`
	Database string `json:"database" envconfig:"DB_NAME" default:"rates_service_db"`
}

// ConnectAndMigrate connects to postgres and applies goose migrations
func ConnectAndMigrate(cfg Config, migrationsDir string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)

	var db *sqlx.DB
	var err error

	for i := 0; i < 20; i++ {
		db, err = sqlx.Connect("pgx", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				log.Printf("Successfully connected to DB")
				break
			}
		}
		log.Printf("Connecting to DB (attempt %d/20), err: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to db after retries: %w", err)
	}

	if err := goose.Up(db.DB, migrationsDir); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}
