package main

import (
	"fmt"
	"log/slog"

	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

func connectDB() (*sqlx.DB, error) {
	slog.Info("Reading DB env variables")
	var (
		host     = os.Getenv("DB_host")
		port, _  = strconv.Atoi(os.Getenv("DB_port"))
		user     = os.Getenv("DB_user")
		password = os.Getenv("DB_password")
		dbname   = os.Getenv("DB_dbname")
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	slog.Debug("Opening connection at:" + psqlInfo)

	db, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}
	slog.Info("Connection opened successfully!")
	return db, nil
}

func runMigrations(db *sqlx.DB) error {
	slog.Info("Initializing db driver")
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize db driver: %w", err)
	}
	slog.Info("DB driver initialized successfully!")

	slog.Info("Migrating")
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}

	slog.Info("Running migrations")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	slog.Info("Migrations applied successfully!")
	return nil
}

func initDB() {
	slog.Info("Initializing database")
	var err error

	db, err = connectDB()
	if err != nil {
		slog.Error("Failed to initialize database connection", "error", err)
		return
	}

	if err := runMigrations(db); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return
	}

	slog.Info("Database initialized successfully!")
}
