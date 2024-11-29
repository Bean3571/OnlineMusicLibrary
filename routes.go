package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func setupRoutes() *mux.Router {
	slog.Info("Initializing router")
	r := mux.NewRouter()
	slog.Info("Router initialized successfully!")

	r.HandleFunc("/songs", addSong).Methods("POST")
	r.HandleFunc("/info", getSongInfo).Methods("GET")
	r.HandleFunc("/songs", updateSong).Methods("PUT")
	r.HandleFunc("/songs", deleteSong).Methods("DELETE")
	r.HandleFunc("/songs/text", getSongText).Methods("GET")
	r.HandleFunc("/songs", getSongsFiltered).Methods("GET")

	r.PathPrefix("/OnlineMusicLibrary/docs/").Handler(http.StripPrefix("/OnlineMusicLibrary/docs/", http.FileServer(http.Dir("docs/"))))
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/OnlineMusicLibrary/docs/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods("GET")

	return r
}

// @Summary Add a new song
// @Description Add a new song by providing group name and song name. Details (release date, text, link) are fetched from an external API.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param input body SongShort true "Group and Song names"
// @Success 201 {string} string "Song added successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Song details not found"
// @Failure 500 {string} string "Failed to add song"
// @Router /songs [post]
func addSong(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to addSong")

	apiURL := os.Getenv("EXTERNAL_API_URL")
	if apiURL == "" {
		slog.Error("External API URL not configured")
		//http.Error(w, "External API URL not configured", http.StatusInternalServerError)
		//return
	}
	slog.Debug("External API URL:", "url", apiURL)

	var input SongShort
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("Invalid JSON format", "error", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if input.GroupName == "" || input.SongName == "" {
		slog.Warn("Missing required fields", "input", input)
		http.Error(w, "Group name and Song name are required", http.StatusBadRequest)
		return
	}

	var songDetail SongDetail
	songDetail.ReleaseDate = "2024-11-25"
	songDetail.Lyrics = "Text Placeholder Verse1 \n\nText Placeholder Verse2\n\nText Placeholder Verse3"
	songDetail.Link = "Link Placeholder"

	fullURL := fmt.Sprintf("%s?group=%s&song=%s",
		apiURL,
		url.QueryEscape(input.GroupName),
		url.QueryEscape(input.SongName),
	)

	resp, err := http.Get(fullURL)
	if err != nil {
		slog.Warn("Failed to call external API, using placeholders", "error", err)
	} else if resp.StatusCode != http.StatusOK {
		slog.Warn("External API returned non-200 status", "status", resp.StatusCode)
		if resp.StatusCode == http.StatusNotFound {
			http.Error(w, "Song details not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Failed to fetch song details", http.StatusInternalServerError)
			return
		}
	} else {
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
			slog.Warn("Failed to decode API response, using placeholders", "error", err)
		}
	}

	var groupID int
	err = db.QueryRow("SELECT id_group FROM musicGroups WHERE groupName = $1", input.GroupName).Scan(&groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = db.QueryRow(
				"INSERT INTO musicGroups (groupName) VALUES ($1) RETURNING id_group",
				input.GroupName,
			).Scan(&groupID)
			if err != nil {
				slog.Error("Failed to insert group", "error", err)
				http.Error(w, "Failed to add song", http.StatusInternalServerError)
				return
			}
		} else {
			slog.Error("Database error while fetching group ID", "error", err)
			http.Error(w, "Failed to add song", http.StatusInternalServerError)
			return
		}
	}

	query := `
        INSERT INTO songs (id_group, song, release_date, lyrics, link)
        VALUES ($1, $2, $3, $4, $5)`
	_, err = db.Exec(query, groupID, input.SongName, songDetail.ReleaseDate, songDetail.Lyrics, songDetail.Link)
	if err != nil {
		slog.Error("Failed to insert song", "error", err)
		http.Error(w, "Failed to add song", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Song added successfully"))

	slog.Info("Song added successfully", "group", input.GroupName, "song", input.SongName)
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
	slog.Info("Received request to get song info")
	groupName := r.URL.Query().Get("group")
	songName := r.URL.Query().Get("song")

	if groupName == "" || songName == "" {
		slog.Warn("Missing required parameters 'groupName' or 'song'")
		http.Error(w, "Missing required parameters 'groupName' or 'song'", http.StatusBadRequest)
		return
	}

	slog.Debug("Fetching song info", "group", groupName, "song", songName)

	var detail SongDetail
	query := `
		SELECT s.release_date, s.lyrics, s.link 
		FROM songs s
		JOIN musicGroups g ON s.id_group = g.id_group
		WHERE g.groupName = $1 AND s.song = $2`
	err := db.Get(&detail, query, groupName, songName)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Song not found", "group", groupName, "song", songName)
			http.Error(w, "Song not found", http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch song details", "error", err)
		http.Error(w, "Failed to fetch song details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)

	slog.Debug("Song details fetched successfully", "group", groupName, "song", songName)
}

// @Summary Update song details
// @Description Update the details of an existing song.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param id_song query int true "ID of the song"
// @Param song body Song true "Updated song details"
// @Success 200 {object} Song "The updated song"
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Failed to update song"
// @Router /songs [put]
func updateSong(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to update song")

	idStr := r.URL.Query().Get("id_song")
	if idStr == "" {
		slog.Warn("Missing required parameter 'id_song'")
		http.Error(w, "Missing required parameter 'id_song'", http.StatusBadRequest)
		return
	}

	idSong, err := strconv.Atoi(idStr)
	if err != nil || idSong < 1 {
		slog.Warn("Invalid 'id_song' parameter", "id_song", idStr)
		http.Error(w, "Invalid 'id_song' parameter", http.StatusBadRequest)
		return
	}

	var song Song
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		slog.Warn("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
        UPDATE songs
        SET 
            id_group = COALESCE($1, id_group),
            song = COALESCE($2, song),
            release_date = COALESCE($3, release_date),
            lyrics = COALESCE($4, lyrics),
            link = COALESCE($5, link)
        WHERE id_song = $6`

	result, err := db.Exec(query,
		song.GroupID,
		song.SongName,
		song.ReleaseDate,
		song.Lyrics,
		song.Link,
		idSong)

	if err != nil {
		slog.Error("Failed to update song details", "error", err)
		http.Error(w, "Failed to update song details", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Warn("Song not found", "id_song", idSong)
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

	slog.Debug("Song details updated successfully", "id_song", idSong)
}

// @Summary Delete a song
// @Description Delete a song from the database by its ID.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param id_song query int true "ID of the song"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Failed to delete song"
// @Router /songs [delete]
func deleteSong(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request deleteSong")

	idStr := r.URL.Query().Get("id_song")
	if idStr == "" {
		slog.Warn("Missing required parameter 'id_song'")
		http.Error(w, "Missing required parameter 'id_song'", http.StatusBadRequest)
		return
	}

	idSong, err := strconv.Atoi(idStr)
	if err != nil || idSong < 1 {
		slog.Warn("Invalid 'id_song' parameter", "id_song", idStr)
		http.Error(w, "Invalid 'id_song' parameter", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM songs WHERE id_song = $1`
	result, err := db.Exec(query, idSong)
	if err != nil {
		slog.Error("Failed to delete song", "error", err)
		http.Error(w, "Failed to delete song", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Warn("Song not found", "id_song", idSong)
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	slog.Debug("Song deleted successfully", "id_song", idSong)
}

// @Summary Get song text with pagination
// @Description Fetch the song text with pagination by verses.
// @Tags songs
// @Accept  json
// @Produce  text/plain
// @Param id_song query int true "ID of the song"
// @Param page query int false "Page number (default is 1)"
// @Param limit query int false "Number of verses per page (default is 1)"
// @Success 200 {string} string "Song text or a portion of it"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 404 {string} string "Song not found or no text available"
// @Failure 500 {string} string "Failed to fetch song text"
// @Router /songs/text [get]
func getSongText(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request getSongText")

	idStr := r.URL.Query().Get("id_song")
	if idStr == "" {
		slog.Warn("Missing required parameter 'id_song'")
		http.Error(w, "Missing required parameter 'id_song'", http.StatusBadRequest)
		return
	}

	idSong, err := strconv.Atoi(idStr)
	if err != nil || idSong < 1 {
		slog.Warn("Invalid 'id_song' parameter", "id_song", idStr)
		http.Error(w, "Invalid 'id_song' parameter", http.StatusBadRequest)
		return
	}

	slog.Debug("Fetching song text", "id_song", idSong)

	var text string
	query := `SELECT lyrics FROM songs WHERE id_song = $1`
	err = db.Get(&text, query, idSong)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Song not found", "id_song", idSong)
			http.Error(w, "Song not found", http.StatusNotFound)
		} else {
			slog.Error("Failed to fetch song text", "error", err)
			http.Error(w, "Failed to fetch song text", http.StatusInternalServerError)
		}
		return
	}

	if text == "" {
		slog.Warn("No text available for song", "id_song", idSong)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("No text available for this song"))
		return
	}

	verses := strings.Split(text, "\n\n")
	page, limit := 1, 1

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			slog.Warn("Invalid page parameter", "page", pageStr)
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	limitStr := r.URL.Query().Get("limit")
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

	slog.Debug("Song text fetched successfully", "id_song", idSong, "page", page, "limit", limit)
}

// @Summary Get songs with optional filters and pagination
// @Description Retrieve a list of songs with optional filters: group name, song name, release date, text, link.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group query string false "Filter by group name"
// @Param song query string false "Filter by song name"
// @Param release_date query string false "Filter by release date (YYYY-MM-DD)"
// @Param text query string false "Filter by text"
// @Param link query string false "Filter by link"
// @Success 200 {array} Song "Filtered list of songs"
// @Failure 500 {string} string "Failed to fetch songs"
// @Router /songs [get]
func getSongsFiltered(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to getSongsWithFilters")

	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")
	releaseDate := r.URL.Query().Get("release_date")
	text := r.URL.Query().Get("text")
	link := r.URL.Query().Get("link")

	query := `
        SELECT s.id_song, s.id_group, g.groupName AS group, s.song, s.release_date, s.lyrics, s.link
        FROM songs s
        INNER JOIN musicGroups g ON s.id_group = g.id_group
        WHERE 1=1`
	args := []interface{}{}

	if group != "" {
		query += " AND g.groupName ILIKE $1"
		args = append(args, "%"+group+"%")
	}
	if song != "" {
		query += " AND s.song ILIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+song+"%")
	}
	if releaseDate != "" {
		query += " AND s.release_date = $" + strconv.Itoa(len(args)+1)
		args = append(args, releaseDate)
	}
	if text != "" {
		query += " AND s.lyrics ILIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+text+"%")
	}
	if link != "" {
		query += " AND s.link ILIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+link+"%")
	}

	slog.Debug("Executing query", "query", query, "args", args)

	var songs []Song
	err := db.Select(&songs, query, args...)
	if err != nil {
		slog.Error("Failed to fetch songs", "error", err)
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)

	slog.Debug("Songs retrieved successfully", "count", len(songs))
}
