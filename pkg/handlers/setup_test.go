package handlers

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/amiranbari/bookings/pkg/config"
	"github.com/amiranbari/bookings/pkg/models"
	"github.com/amiranbari/bookings/pkg/renders"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "templates"
var functions = template.FuncMap{
	"humanDate":  renders.HumanDate,
	"formatDate": renders.FormatDate,
	"iterate":    renders.Iterate,
}

func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}

func TestMain(m *testing.M) {
	//Say what we need to put in out session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	// change this to true in production
	app.InProduction = false

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(app.MailChan)

	listenForMail()

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = tc
	app.UseCache = true

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	renders.NewRenderer(&app)

	os.Exit(m.Run())
}

func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

//Loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func CreateTestTemplateCache() (config.TemplateCache, error) {
	myCache := config.TemplateCache{}

	pages, err := filepath.Glob(fmt.Sprintf("../../%s/*.page.html", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("../../%s/*.layout.html", pathToTemplates))

		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("../../%s/*.layout.html", pathToTemplates))

			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}

func getRoutes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/json", Repo.Json)

	mux.Get("/reservation", Repo.Reservation)

	//search
	mux.Get("/search", Repo.Search)
	mux.Post("/search", Repo.PostSearch)
	mux.Get("/choose-room/{id}", Repo.ChooseRoom)
	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostReservation)

	//user
	mux.Get("/login", Repo.Login)
	mux.Post("/login", Repo.PostLogin)
	mux.Get("/logout", Repo.Logout)

	mux.Get("/admin/dashboard", Repo.Dashboard)
	mux.Get("/admin/reservations", Repo.AdminReservations)
	mux.Get("/admin/new-reservations", Repo.AdminNewReservations)
	mux.Get("/admin/reservations/{id}", Repo.AdminShowReservations)
	mux.Post("/admin/reservations/{id}", Repo.AdminPostShowReservations)
	mux.Get("/admin/reservations/{id}/processed", Repo.AdminPutShowReservations)
	mux.Get("/admin/reservations/{id}/delete", Repo.AdminDeleteReservation)
	mux.Get("/admin/reservations-calender", Repo.AdminReservationsCalender)
	mux.Post("/admin/reservations-calender", Repo.AdminPostReservationsCalender)

	fileServer := http.FileServer(http.Dir("../../static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
