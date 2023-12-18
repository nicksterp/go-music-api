package main

/*
Sets up environment variables, database connection, and routes
*/

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func main() {

	err := loadEnv(".env")
	if err != nil {
		log.Fatalf("Error loading .env: %v", err)
	}

	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Route("/song", func(r chi.Router) {
		r.Get("/", getSong(db))
		r.With(TokenAuthMiddleware).Post("/", createSong(db))
		r.Get("/history", getSongHistory(db))
		r.Post("/submit", submitSong(db))
	})

}
