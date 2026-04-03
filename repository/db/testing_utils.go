package db

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgres(t *testing.T, cfg *Config) (*postgres.PostgresContainer, string, error) {
	ctx := t.Context()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, "", err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", err
	}

	return pgContainer, connStr, nil
}

func ShutdownContainer(t *testing.T, container *postgres.PostgresContainer) error {
	return container.Terminate(t.Context())
}

func SetupPGFlow(s *suite.Suite, migrationsPath string) (*postgres.PostgresContainer, *sqlx.DB, string) {
	container, dsn, err := SetupPostgres(s.T(), defaultPGConfigs())
	s.Require().NoError(err)

	conn, err := sqlx.Connect("pgx", dsn)
	s.Require().NoError(err)

	err = goose.UpContext(s.T().Context(), conn.DB, migrationsPath)
	s.Require().NoError(err)

	return container, conn, dsn
}

func TearDownPGFlow(s *suite.Suite, container *postgres.PostgresContainer, conn *sqlx.DB, dsn, migrationsPath string) {
	err := goose.DownContext(s.T().Context(), conn.DB, migrationsPath)
	s.Require().NoError(err)

	err = conn.Close()
	s.Require().NoError(err)

	err = ShutdownContainer(s.T(), container)
	s.Require().NoError(err)
}

func defaultPGConfigs() *Config {
	return &Config{
		Host:     "host",
		Port:     5432,
		User:     "user",
		Password: "password",
		Database: "name",
	}
}
