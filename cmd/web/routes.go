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
	mux.Get("/json", handlers.Repo.Json)

	mux.Get("/reservation", handlers.Repo.Reservation)

	//search
	mux.Get("/search", handlers.Repo.Search)
	mux.Post("/search", handlers.Repo.PostSearch)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/make-reservation", handlers.Repo.MakeReservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)

	//user
	mux.Get("/login", handlers.Repo.Login)
	mux.Post("/login", handlers.Repo.PostLogin)
	mux.Get("/logout", handlers.Repo.Logout)

	//admin dashboard
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.Dashboard)
		mux.Get("/reservations", handlers.Repo.AdminReservations)
		mux.Get("/new-reservations", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations/{id}", handlers.Repo.AdminShowReservations)
		mux.Post("/reservations/{id}", handlers.Repo.AdminPostShowReservations)
		mux.Get("/reservations/{id}/processed", handlers.Repo.AdminPutShowReservations)
	})

	fileServer := http.FileServer(http.Dir("../../static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}
