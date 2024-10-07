package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func setupRoutes() *mux.Router {
	slog.Info("Initializing router")
	r := mux.NewRouter()
	slog.Info("Router initialized successfully!")

	slog.Info("Setting up /songs routes")
	r.HandleFunc("/songs", addSong).Methods("POST")
	r.HandleFunc("/info", getSongInfo).Methods("GET")
	r.HandleFunc("/songs", updateSong).Methods("PUT")
	r.HandleFunc("/songs", deleteSong).Methods("DELETE")
	r.HandleFunc("/songs/text", getSongText).Methods("GET")
	r.HandleFunc("/songs", getSongs).Methods("GET")
	slog.Info("/songs Routes set up successfully!")

	slog.Info("Setting up Swagger route")
	r.PathPrefix("/OnlineMusicLibrary/docs/").Handler(http.StripPrefix("/OnlineMusicLibrary/docs/", http.FileServer(http.Dir("docs/"))))
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/OnlineMusicLibrary/docs/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods("GET")
	slog.Info("Swagger route set up successfully!")

	return r
}

// @Summary Add a new song
// @Description Add a song with only group and song fields, or all fields.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param song body SongShort true "Song data"
// @Success 201 {object} Song
// @Failure 400 {string} string "Invalid JSON format"
// @Failure 500 {string} string "Failed to add song"
// @Router /songs [post]
func addSong(w http.ResponseWriter, r *http.Request) {
	slog.Info("Recieved request addSong")
	var song Song

	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		slog.Error("Invalid JSON format", "error", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	slog.Debug("Adding new song", "group", song.Group, "song", song.Song)

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM songs WHERE "group" = $1 AND "song" = $2)`
	err := db.Get(&exists, checkQuery, song.Group, song.Song)
	if err != nil {
		slog.Error("Failed to check if song exists", "error", err)
		http.Error(w, "Failed to check if song exists", http.StatusInternalServerError)
		return
	}

	if exists {
		slog.Warn("Song already exists", "group", song.Group, "song", song.Song)
		http.Error(w, "Song already exists", http.StatusConflict)
		return
	}

	if song.Text == "" {
		song.Text = "Text placeholder"
	}
	if song.ReleaseDate == "" {
		song.ReleaseDate = "ReleaseDate placeholder"
	}
	if song.Link == "" {
		song.Link = "Link placeholder"
	}

	query := `INSERT INTO songs ("group", "song", "text", "release_date", "link")
                 VALUES ($1, $2, $3, $4, $5)`
	slog.Debug("Executing query:" + query)
	_, err = db.Exec(query, song.Group, song.Song, song.Text, song.ReleaseDate, song.Link)

	if err != nil {
		slog.Error("Failed to add song", "error", err)
		http.Error(w, "Failed to add song", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(song)

	slog.Debug("Song added successfully", "group", song.Group, "song", song.Song)
}

// @Summary Music info
// @Description Get releaseDate, text, link for a song based on group and song.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group query string true "Group of the song"
// @Param song query string true "Title of the song"
// @Success 200 {object} SongDetail "Song details"
// @Failure 400 {string} string "Missing required parameters 'group' or 'song'"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Failed to fetch song details"
// @Router /info [get]
func getSongInfo(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request getSongInfo")
	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")

	if group == "" || song == "" {
		slog.Warn("Missing required parameters 'group' or 'song'")
		http.Error(w, "Missing required parameters 'group' or 'song'", http.StatusBadRequest)
		return
	}

	slog.Debug("Fetching song info", "group", group, "song", song)

	var detail SongDetail
	query := `SELECT release_date, text, link FROM songs WHERE "group" = $1 AND "song" = $2`
	err := db.Get(&detail, query, group, song)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Song not found", "group", group, "song", song)
			http.Error(w, "Song not found", http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch song details", "error", err)
		http.Error(w, "Failed to fetch song details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)

	slog.Debug("Song details fetched successfully", "group", group, "song", song)
}

// @Summary Update a song
// @Description Update a song by providing all fields (group, song, releaseDate, text, link).
// @Tags songs
// @Accept  json
// @Produce  json
// @Param song body Song true "Song data to update"
// @Success 200 {string} string "Song updated successfully"
// @Failure 400 {string} string "All fields (group, song, releaseDate, text, link) are required"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Failed to update song"
// @Router /songs [put]
func updateSong(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request updateSong")
	var song Song

	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		slog.Error("Invalid JSON format", "error", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if song.Group == "" || song.Song == "" || song.ReleaseDate == "" || song.Text == "" || song.Link == "" {
		slog.Warn("Missing required fields", "group", song.Group, "song", song.Song)
		http.Error(w, "All fields (group, song, releaseDate, text, link) are required", http.StatusBadRequest)
		return
	}

	slog.Debug("Updating song", "group", song.Group, "song", song.Song)

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM songs WHERE "group" = $1 AND "song" = $2)`
	err := db.Get(&exists, checkQuery, song.Group, song.Song)
	if err != nil {
		slog.Error("Failed to check if song exists", "error", err)
		http.Error(w, "Failed to check if song exists", http.StatusInternalServerError)
		return
	}

	if !exists {
		slog.Warn("Song not found", "group", song.Group, "song", song.Song)
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	updateQuery := `UPDATE songs
                    SET "text" = $1, "release_date" = $2, "link" = $3
                    WHERE "group" = $4 AND "song" = $5`
	_, err = db.Exec(updateQuery, song.Text, song.ReleaseDate, song.Link, song.Group, song.Song)
	if err != nil {
		slog.Error("Failed to update song", "error", err)
		http.Error(w, "Failed to update song", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Song updated successfully"))

	slog.Debug("Song updated successfully", "group", song.Group, "song", song.Song)
}

// @Summary Delete a song
// @Description Delete a song by specifying the group and song fields.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param song body SongShort true "Song data to delete"
// @Success 200 {string} string "Song deleted successfully"
// @Failure 400 {string} string "Invalid JSON format"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Failed to delete song"
// @Router /songs [delete]
func deleteSong(w http.ResponseWriter, r *http.Request) {
	slog.Info("Recieved request deleteSong")
	var song Song

	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		slog.Error("Invalid JSON format", "error", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	slog.Debug("Deleting song", "group", song.Group, "song", song.Song)

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM songs WHERE "group" = $1 AND "song" = $2)`
	err := db.Get(&exists, checkQuery, song.Group, song.Song)
	if err != nil {
		slog.Error("Failed to check if song exists", "error", err)
		http.Error(w, "Failed to check if song exists", http.StatusInternalServerError)
		return
	}

	if !exists {
		slog.Warn("Song not found", "group", song.Group, "song", song.Song)
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	deleteQuery := `DELETE FROM songs WHERE "group" = $1 AND "song" = $2`
	_, err = db.Exec(deleteQuery, song.Group, song.Song)
	if err != nil {
		slog.Error("Failed to delete song", "error", err)
		http.Error(w, "Failed to delete song", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Song deleted successfully"))

	slog.Debug("Song deleted successfully", "group", song.Group, "song", song.Song)
}

// @Summary Get song text with pagination
// @Description Fetch the song text with pagination by verses.
// @Tags songs
// @Accept  json
// @Produce  text/plain
// @Param group query string true "Group of the song"
// @Param song query string true "Title of the song"
// @Param page query int false "Page number (default is 1)"
// @Param limit query int false "Number of verses per page (default is 1)"
// @Success 200 {string} string "Song text or a portion of it"
// @Failure 400 {string} string "Invalid page or limit parameter or missing required parameters"
// @Failure 404 {string} string "Song not found or no text available"
// @Failure 500 {string} string "Failed to fetch song text"
// @Router /songs/text [get]
func getSongText(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request getSongText")

	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")

	if group == "" || song == "" {
		slog.Warn("Missing required parameters 'group' or 'song'")
		http.Error(w, "Missing required parameters 'group' or 'song'", http.StatusBadRequest)
		return
	}

	slog.Debug("Fetching song text", "group", group, "song", song)

	var songText string
	query := `SELECT "text" FROM songs WHERE "group" = $1 AND "song" = $2`
	err := db.Get(&songText, query, group, song)
	if err != nil {
		slog.Error("Failed to fetch song text", "error", err)
		http.Error(w, "Failed to fetch song text", http.StatusInternalServerError)
		return
	}

	if songText == "" {
		slog.Warn("Text not found for song", "group", group, "song", song)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Looks like we don't have text for this one"))
		return
	}

	verses := strings.Split(songText, "\n\n")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 1

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			slog.Warn("Invalid page parameter", "page", pageStr)
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			slog.Warn("Invalid limit parameter", "limit", limitStr)
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= len(verses) {
		slog.Warn("Page out of range", "page", page)
		http.Error(w, "Page out of range", http.StatusBadRequest)
		return
	}

	if end > len(verses) {
		end = len(verses)
	}

	paginatedVerses := verses[start:end]
	responseText := strings.Join(paginatedVerses, "\n\n")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseText))

	slog.Debug("Song text fetched successfully", "group", group, "song", song, "page", page, "limit", limit)
}

// @Summary Get songs with filters and pagination
// @Description Get songs with optional filters and pagination.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group       query string false "Group of the song"
// @Param song        query string false "Title of the song"
// @Param releaseDate query string false "Release date of the song"
// @Param link        query string false "Link to the song"
// @Param page        query int    false "Page number (default is 1)"
// @Param limit       query int    false "Number of songs per page (default is 3)"
// @Success 200 {array}  Song   "List of songs"
// @Failure 400 {string} string "Invalid page or limit parameter"
// @Failure 404 {string} string "No songs found"
// @Failure 500 {string} string "Failed to fetch songs"
// @Router /songs [get]
func getSongs(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request getSongs")

	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")
	releaseDate := r.URL.Query().Get("releaseDate")
	link := r.URL.Query().Get("link")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 3

	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			slog.Warn("Invalid page parameter", "page", pageStr)
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			slog.Warn("Invalid limit parameter", "limit", limitStr)
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	query := `SELECT "group", "song", "release_date", "text", "link" FROM songs WHERE 1=1`

	var args []interface{}
	argIndex := 1

	if group != "" {
		query += fmt.Sprintf(` AND "group" = $%d`, argIndex)
		args = append(args, group)
		argIndex++
	}

	if song != "" {
		query += fmt.Sprintf(` AND "song" = $%d`, argIndex)
		args = append(args, song)
		argIndex++
	}

	if releaseDate != "" {
		query += fmt.Sprintf(` AND "release_date" = $%d`, argIndex)
		args = append(args, releaseDate)
		argIndex++
	}

	if link != "" {
		query += fmt.Sprintf(` AND "link" = $%d`, argIndex)
		args = append(args, link)
		argIndex++
	}

	offset := (page - 1) * limit
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	args = append(args, limit, offset)

	var songs []Song
	err = db.Select(&songs, query, args...)
	if err != nil {
		slog.Error("Failed to fetch songs", "error", err)
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		return
	}

	if len(songs) == 0 {
		slog.Warn("No songs found with the given filters")
		http.Error(w, "No songs found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)

	slog.Debug("Songs fetched successfully", "total_songs", len(songs), "page", page, "limit", limit)
}
