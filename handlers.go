package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
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

type SongRequest struct {
	SongLink string `json:"songLink"`
}

// HANDLERS

/*
GET /song
Returns the most recently recommended song
*/
func getSong(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var song Song

		err := db.QueryRow("SELECT * FROM song ORDER BY id DESC LIMIT 1").Scan(&song.ID, &song.Title, &song.Artist, &song.ImageURL, &song.SubmittedAt, &song.SongURL, &song.Platform)

		if err == sql.ErrNoRows {
			render.Render(w, r, ErrNotFound)
			return
		}

		if err != nil {
			log.Println(err)
			render.Render(w, r, ErrServer)
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

Headers: "Authorization: API_TOKEN"
Body: {"songLink": "SPOTIFY_SONG_LINK"}
*/
func createSong(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var songReq SongRequest
		err := json.NewDecoder(r.Body).Decode(&songReq)
		if err != nil {
			log.Println("Error decoding song request:", err)
			render.Render(w, r, ErrServer)
			return
		}

		// Check provider (Soundcloud, Spotify, Youtube) from the link and parse accordingly
		if strings.Contains(songReq.SongLink, "soundcloud.com") {
			// UNSUPPORTED: Soundcloud not releasing API keys at this time
			render.Render(w, r, ErrInvalidRequest(errors.New("unsupported provider")))
			return
		} else if strings.Contains(songReq.SongLink, "spotify.com") {
			// Parse link for song ID
			u, err := url.Parse(songReq.SongLink)
			if err != nil {
				log.Println("Error parsing song link:", err)
				render.Render(w, r, ErrServer)
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
				log.Println("Error creating Spotify API request:", err)
				render.Render(w, r, ErrServer)
				return
			}

			// Add Authorization header to the request
			refreshSpotifyToken()
			req.Header.Add("Authorization", "Bearer "+os.Getenv("SPOTIFY_ACCESS_TOKEN"))

			// Send the request and get the response
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Println("Error sending Spotify API request:", err)
				render.Render(w, r, ErrServer)
				return
			}
			defer resp.Body.Close()

			// Decode the response body into a map
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				log.Println("Error decoding Spotify API response:", err)
				render.Render(w, r, ErrServer)
				return
			}

			var title, artist, imageURL string

			if name, ok := result["name"].(string); ok {
				title = name
			}

			if artists, ok := result["artists"].([]interface{}); ok && len(artists) > 0 {
				if artistData, ok := artists[0].(map[string]interface{}); ok {
					if name, ok := artistData["name"].(string); ok {
						artist = name
					}
				}
			}

			if album, ok := result["album"].(map[string]interface{}); ok {
				if images, ok := album["images"].([]interface{}); ok && len(images) > 0 {
					if imageData, ok := images[0].(map[string]interface{}); ok {
						if url, ok := imageData["url"].(string); ok {
							imageURL = url
						}
					}
				}
			}

			song := Song{
				ID:          nil, // Make sure this is the intended behavior
				Title:       title,
				Artist:      artist,
				ImageURL:    imageURL,
				SubmittedAt: time.Now().Truncate(24 * time.Hour),
				SongURL:     songReq.SongLink, // Ensure songReq.SongLink is defined and valid
				Platform:    "Spotify",
			}

			// Prepare SQL statement
			stmt, err := db.Prepare("INSERT INTO Song(title, artist, image_url, submitted_at, song_url, platform) VALUES($1, $2, $3, $4, $5, $6)")
			if err != nil {
				log.Println("Error preparing SQL statement:", err)
				render.Render(w, r, ErrServer)
				return
			}
			defer stmt.Close()

			// Execute SQL statement
			_, err = stmt.Exec(song.Title, song.Artist, song.ImageURL, song.SubmittedAt, song.SongURL, song.Platform)
			if err != nil {
				log.Println("Error executing SQL statement:", err)
				render.Render(w, r, ErrServer)
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
Allows user to submit a song recommendation, requires solved captcha
*/
func submitSong(*sql.DB) http.HandlerFunc { //TODO: Implement
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
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
var ErrServer = &ErrResponse{HTTPStatusCode: 500, StatusText: "Server error."}

// SPOTIFY AUTH TOKEN HELPER
func refreshSpotifyToken() {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	tokenURL := "https://accounts.spotify.com/api/token"

	// Check if token is already set and still valid
	if token, expiry := os.Getenv("SPOTIFY_ACCESS_TOKEN"), os.Getenv("SPOTIFY_TOKEN_EXPIRY"); token != "" && expiry != "" {
		expiryTime, err := time.Parse(time.RFC3339, expiry)
		if err == nil && time.Now().Before(expiryTime) {
			return
		}
	}

	// Request new token
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		log.Println("Error creating Spotify API request:", err)
		return
	}
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending Spotify API request:", err)
		return
	}
	defer resp.Body.Close()

	// Decode the response body into a map
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println("Error decoding Spotify API response:", err)
		return
	}

	// Set new token and expiry
	if accessToken, ok := result["access_token"].(string); ok {
		os.Setenv("SPOTIFY_ACCESS_TOKEN", accessToken)
		if expiresIn, ok := result["expires_in"].(float64); ok {
			expiry := time.Now().Add(time.Duration(expiresIn) * time.Second).Format(time.RFC3339)
			os.Setenv("SPOTIFY_TOKEN_EXPIRY", expiry)
		}
	}
}
