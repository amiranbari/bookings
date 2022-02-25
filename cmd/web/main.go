package main

import (
	// "errors"
	"encoding/gob"
	"fmt"
	"github.com/amiranbari/bookings/internal/driver"
	"github.com/amiranbari/bookings/internal/helpers"
	"log"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/amiranbari/bookings/pkg/config"
	"github.com/amiranbari/bookings/pkg/handlers"
	"github.com/amiranbari/bookings/pkg/models"
	"github.com/amiranbari/bookings/pkg/renders"

	// "math/rand"
	"net/http"
	// "time"
)

// var randomSource = rand.NewSource(time.Now().Unix())
// var random = rand.New(randomSource)

const portNumber string = ":8000"

// func makeRandomNumber() int {
// 	return random.Intn(100)
// }

// func devide(rw http.ResponseWriter, r *http.Request) {
// 	var x float32 = float32(makeRandomNumber())
// 	var y float32 = float32(makeRandomNumber())
// 	f, err := devideValues(x, y)
// 	if err != nil {
// 		fmt.Fprintf(rw, "Cannot devide by zero")
// 		return
// 	}
// 	fmt.Fprintf(rw, fmt.Sprintf("%f devided by %f is %f", x, y, f))
// }

// func devideValues(x, y float32) (float32, error) {
// 	if y <= 0 {
// 		return 0, errors.New("cannot devide zero")
// 	}

// 	return x / y, nil
// }

// func addValues(x, y int) int {
// 	return x + y
// }

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {

	db, err := run()

	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	// http.HandleFunc("/", handlers.Repo.Home)
	// http.HandleFunc("/about", handlers.Repo.About)

	// http.HandleFunc("/devide", devide)

	fmt.Println(fmt.Sprintf("starting application on port number %s", portNumber))

	// _ = http.ListenAndServe(portNumber, nil)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: route(&app),
	}

	err = srv.ListenAndServe()

	log.Fatal(err)
}

func run() (*driver.DB, error) {
	//Say what we need to put in out session
	gob.Register(models.Reservation{})

	// change this to true in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	//connect to database
	log.Println("Connecting to database ...")
	db, err := driver.ConnectSql("host=localhost port=5432 dbname=test user=postgres password=123456")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying ...")
	}

	log.Println("Connected to database!")

	tc, err := renders.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	renders.NewTemplates(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
