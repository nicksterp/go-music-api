package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
)

// STRUCTS
type Song struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	ImageURL    string `json:"image_url"`
	SubmittedAt string `json:"submitted_at"`
	SongURL     string `json:"song_url"`
	Platform    string `json:"platform"`
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
*/
func createSong(*sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var songLink string
		err := json.NewDecoder(r.Body).Decode(&songLink)
		if err != nil {
			render.Render(w, r, ErrRender(err))
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
