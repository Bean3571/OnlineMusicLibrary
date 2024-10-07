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

func initDB() {
	var err error

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
	db, err = sqlx.Open("postgres", psqlInfo)
	if err != nil {
		slog.Error("Failed to connect to the database: " + err.Error())
	}
	slog.Info("Connection opened successfully!")

	slog.Info("Initializing db driver")
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("DB driver initialized successfully!")

	slog.Info("Migrating")
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("Migrated successfully!")

	slog.Info("Creating tables")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error(err.Error())
	}
	slog.Info("Tables created successfully!")
}
