package main

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

}
