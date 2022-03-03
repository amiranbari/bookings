package handlers

import (
	"encoding/json"
	"github.com/amiranbari/bookings/internal/driver"
	"github.com/amiranbari/bookings/internal/helpers"
	"github.com/amiranbari/bookings/internal/repository"
	"github.com/amiranbari/bookings/internal/repository/dbrepo"
	"net/http"
	"strconv"
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

func (m *Repository) PostHome(rw http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(rw, err)
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

	roomId, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("firstname"),
		LastName:  r.Form.Get("lastname"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomId:    roomId,
	}

	form := forms.New(r.PostForm)
	form.Has("firstname")
	form.Has("lastname")
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		renders.Template(rw, r, "home.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	newReservationId, err := m.DB.InsertReservation(reservation)

	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	restriction := models.RoomRestriction{
		RoomId:        roomId,
		ReservationId: newReservationId,
		RestrictionId: 1,
		StartDate:     startDate,
		EndDate:       endDate,
	}

	err = m.DB.InsertRoomRestriction(restriction)

	if err != nil {
		helpers.ServerError(rw, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(rw, r, "/reservation", http.StatusSeeOther)

}

func (m *Repository) Reservation(rw http.ResponseWriter, r *http.Request) {

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	renders.Template(rw, r, "reservation.page.html", &models.TemplateData{
		Data: data,
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
		helpers.ServerError(rw, err)
		return
	}

	form := forms.New(r.PostForm)
	form.Has("start_date")
	form.Has("end_date")

	if !form.Valid() {
		renders.Template(rw, r, "search.page.html", &models.TemplateData{})
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
		helpers.ServerError(rw, err)
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
