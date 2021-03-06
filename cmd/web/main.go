package main

import (
	// "errors"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/amiranbari/bookings/internal/driver"
	"github.com/amiranbari/bookings/internal/helpers"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
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

	defer close(app.MailChan)
	fmt.Println("Starting mail listening ...")
	listenForMail()

	fmt.Println(fmt.Sprintf("starting application on port number %s", portNumber))

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
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	inProduction, _ := strconv.ParseBool(os.Getenv("PRODUCTION"))

	useCache := flag.Bool("cache", false, "User cache for templates or not!")
	flag.Parse()

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// change this to true in production
	app.InProduction = inProduction

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
	db, err := driver.ConnectSql("host=localhost port=5432 dbname=bookings user=postgres password=123456")
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
	app.UseCache = *useCache

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	renders.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
