package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/amiranbari/bookings/internal/driver"
	"github.com/amiranbari/bookings/internal/helpers"
	"github.com/amiranbari/bookings/internal/repository"
	"github.com/amiranbari/bookings/internal/repository/dbrepo"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amiranbari/bookings/internal/forms"
	"github.com/amiranbari/bookings/pkg/config"
	"github.com/amiranbari/bookings/pkg/models"
	"github.com/amiranbari/bookings/pkg/renders"
)

//Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

//NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

//NewTestRepo creates a new testing repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

//NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(rw http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation
	renders.Template(rw, r, "home.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) About(rw http.ResponseWriter, r *http.Request) {
	renders.Template(rw, r, "about.page.html", &models.TemplateData{})
}

func (m *Repository) Reservation(rw http.ResponseWriter, r *http.Request) {

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	renders.Template(rw, r, "reservation.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

type jsonResponse struct {
	OK      bool
	Message string
}

func (m *Repository) Json(rw http.ResponseWriter, r *http.Request) {
	jsonResponse := jsonResponse{
		true,
		"successfully",
	}
	out, err := json.MarshalIndent(jsonResponse, "", "  ")
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(out)
}

func (m *Repository) Search(rw http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation
	renders.Template(rw, r, "search.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) PostSearch(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(rw, r, "/search", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("start_date", "end_date")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "form is not valid!")
		http.Redirect(rw, r, "/search", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	// 2020-01-01 -- 01/02 03:0405PM '06 --700
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't search in availability rooms!")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No available room!")
		http.Redirect(rw, r, "/search", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	renders.Template(rw, r, "choose-room.page.html", &models.TemplateData{
		Data: data,
	})

}

func (m *Repository) ChooseRoom(rw http.ResponseWriter, r *http.Request) {
	// split the URL up by /, and grab the 3rd element
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(rw, r, "/search", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(rw, r, "/search", http.StatusTemporaryRedirect)
		return
	}

	res.RoomId = roomID
	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(rw, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) MakeReservation(rw http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomById(res.RoomId)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room!")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.Title = room.Title
	data["reservation"] = res

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)

	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	m.App.Session.Put(r.Context(), "reservation", res)

	renders.Template(rw, r, "make-reservation.page.html", &models.TemplateData{
		Data:      data,
		Form:      forms.New(nil),
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(rw http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(rw, r, "/make-reservation", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("firstname", "lastname", "email", "phone")
	form.IsEmail("email")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "Form is not valid!")
		http.Redirect(rw, r, "/make-reservation", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(rw, r, "/make-reservation", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("firstname"),
		LastName:  r.Form.Get("lastname"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: res.StartDate,
		EndDate:   res.EndDate,
		RoomId:    res.RoomId,
	}
	reservation.Room.Title = res.Room.Title

	newReservationId, err := m.DB.InsertReservation(reservation)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation to database!")
		http.Redirect(rw, r, "/make-reservation", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		RoomId:        res.RoomId,
		ReservationId: newReservationId,
		RestrictionId: 1,
		StartDate:     res.StartDate,
		EndDate:       res.EndDate,
	}

	err = m.DB.InsertRoomRestriction(restriction)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert restriction to database!")
		http.Redirect(rw, r, "/make-reservation", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	//send reservation mail

	html := fmt.Sprintf(`
		<strong>Reservation Confirmation</stronge><br>
		Dear %s: <br>
		This is to confirm your reservation from %s to %s.
		`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:      reservation.Email,
		From:    "me@here.com",
		Subject: "Reservation confirmation",
		Content: html,
	}

	m.App.MailChan <- msg

	//send room mail
	html = fmt.Sprintf(`
		<strong>Reservation Notification</stronge><br>
		A reservation has been made for %s from %s to %s.
		`, reservation.Room.Title, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:      reservation.Email,
		From:    "me@here.com",
		Subject: "Reservation confirmation",
		Content: html,
	}

	m.App.MailChan <- msg

	http.Redirect(rw, r, "/reservation", http.StatusSeeOther)

}

// Login users
func (m *Repository) Login(rw http.ResponseWriter, r *http.Request) {
	renders.Template(rw, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin users
func (m *Repository) PostLogin(rw http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	form.MinLength("password", 8)
	if !form.Valid() {
		renders.Template(rw, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(rw, r, "/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

// Logout users
func (m *Repository) Logout(rw http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

// Dashboard admin
func (m *Repository) Dashboard(rw http.ResponseWriter, r *http.Request) {
	renders.Template(rw, r, "dashboard.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) AdminReservations(rw http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}
	data["reservations"] = reservations
	renders.Template(rw, r, "admin-reservations.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) AdminNewReservations(rw http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}
	data["reservations"] = reservations
	renders.Template(rw, r, "admin-reservations.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) AdminShowReservations(rw http.ResponseWriter, r *http.Request) {
	//data := make(map[string]interface{})

	exploded := strings.Split(r.RequestURI, "/")
	fmt.Println(exploded)
	//roomID, err := strconv.Atoi(exploded[2])
	//if err != nil {
	//	m.App.Session.Put(r.Context(), "error", "missing url parameter")
	//	http.Redirect(rw, r, "/search", http.StatusTemporaryRedirect)
	//	return
	//}
	//
	//renders.Template(rw, r, "admin-show-reservation.page.html", &models.TemplateData{
	//	Form: forms.New(nil),
	//	Data: data,
	//})
}
