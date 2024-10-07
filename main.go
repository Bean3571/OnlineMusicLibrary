package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

// @title OnlineMusicLibrary API
// @version 1.0
// @description Online Music Library API project.
// @host localhost:8080
// @BasePath /

func main() {
	slog.Info("Starting")

	slog.Info("Formatting logger")
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
		}),
	)
	slog.SetDefault(logger)

	slog.Info("Loading.env")
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file")
	}
	slog.Info(".env loaded successfully!")

	slog.Info("Initializing DB")
	initDB()
	if db == nil {
		slog.Error("Database connection is nil after initialization")
	}
	slog.Info("DB initialized successfully!")

	slog.Info("Setting up routes")
	r := setupRoutes()
	slog.Info("Routes set up successfully!")

	slog.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		slog.Error("Error starting server")
		os.Exit(1)
	}
}