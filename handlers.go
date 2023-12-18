package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/render"
)

// AUTH MIDDLEWARE
func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("API_TOKEN") // Get the token from the environment variable
		token := r.Header.Get("Authorization")

		if token != secret {
			render.Render(w, r, ErrUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// STRUCTS
type Song struct {
	ID          *int      `json:"id"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	ImageURL    string    `json:"image_url"`
	SubmittedAt time.Time `json:"submitted_at"`
	SongURL     string    `json:"song_url"`
	Platform    string    `json:"platform"`
}

// HANDLERS

/*
GET /song
Returns the most recently recommended song
*/
func getSong(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var song Song

		err := db.QueryRow("SELECT * FROM songs ORDER BY id DESC LIMIT 1").Scan(&song.ID, &song.Title, &song.Artist, &song.ImageURL, &song.SubmittedAt, &song.SongURL, &song.Platform)

		if err == sql.ErrNoRows {
			render.Render(w, r, ErrNotFound)
			return
		}

		if err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}

		render.JSON(w, r, song)
	})
}

/*
GET /song/history
Returns all Songs that have been recommended, paginated
*/
func getSongHistory(*sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

/*
POST /song
Creates a new Song recommendation
Only accessible by admins
currently supports only Spotify links
*/
func createSong(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var songLink string
		err := json.NewDecoder(r.Body).Decode(&songLink)
		if err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}

		// Check provider (Soundcloud, Spotify, Youtube) from the link and parse accordingly
		if strings.Contains(songLink, "soundcloud.com") {
			// UNSUPPORTED: Soundcloud not releasing API keys at this time
			render.Render(w, r, ErrInvalidRequest(errors.New("unsupported provider")))
			return
		} else if strings.Contains(songLink, "spotify.com") {
			// Parse link for song ID
			u, err := url.Parse(songLink)
			if err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}
			songID := ""
			pathParts := strings.Split(u.Path, "/")
			for i, part := range pathParts {
				if part == "track" && i+1 < len(pathParts) {
					songID = strings.Split(pathParts[i+1], "?")[0]
				}
			}
			if songID == "" {
				render.Render(w, r, ErrInvalidRequest(errors.New("Song ID not found")))
				return
			}

			// Create a new HTTP request to the Spotify API endpoint for getting track details
			req, err := http.NewRequest("GET", "https://api.spotify.com/v1/tracks/"+songID, nil)
			if err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}

			// Send the request and get the response
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}
			defer resp.Body.Close()

			// Decode the response body into a map
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

			// Create a new Song object and fill the fields with the data from the response
			song := Song{
				ID:          nil,
				Title:       result["name"].(string),
				Artist:      result["artists"].([]interface{})[0].(map[string]interface{})["name"].(string),
				ImageURL:    result["album"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"].(string),
				SubmittedAt: time.Now().Truncate(24 * time.Hour), // Set the current time in RFC3339 format
				SongURL:     songLink,                            // Assuming this is the correct value
				Platform:    "Spotify",                           // Assuming this is the correct value
			}

			// Prepare SQL statement
			stmt, err := db.Prepare("INSERT INTO Song(Title, Artist, ImageURL, SubmittedAt, SongURL, Platform) VALUES(?, ?, ?, ?, ?, ?)")
			if err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}
			defer stmt.Close()

			// Execute SQL statement
			_, err = stmt.Exec(song.Title, song.Artist, song.ImageURL, song.SubmittedAt, song.SongURL, song.Platform)
			if err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}

		} else {
			render.Render(w, r, ErrInvalidRequest(errors.New("invalid provider")))
			return
		}
	})
}

/*
POST /song/submit
Allows user to submit a song recommendation, rate limited
*/
func submitSong(*sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

// ERRORS
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
var ErrUnauthorized = &ErrResponse{HTTPStatusCode: 401, StatusText: "Unauthorized."}
