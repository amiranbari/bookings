package main

import (
	"net/http"

	"github.com/amiranbari/bookings/pkg/config"
	"github.com/amiranbari/bookings/pkg/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func route(app *config.AppConfig) http.Handler {
	// mux := pat.New()

	// mux.Get("/", http.HandlerFunc(handlers.Repo.Home))

	// mux.Get("/about", http.HandlerFunc(handlers.Repo.About))

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Post("/", handlers.Repo.PostHome)
	mux.Get("/json", handlers.Repo.Json)

	mux.Get("/reservation", handlers.Repo.Reservation)

	//search
	mux.Get("/search", handlers.Repo.Search)
	mux.Post("/search", handlers.Repo.PostSearch)

	fileServer := http.FileServer(http.Dir("../../static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
