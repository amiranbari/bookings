package main

import (
	// "errors"
	"encoding/gob"
	"fmt"
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

	err := run()

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

func run() error {
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

	tc, err := renders.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
		return err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	renders.NewTemplates(&app)
	helpers.NewHelpers(&app)

	return nil
}
