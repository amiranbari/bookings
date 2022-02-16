package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

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
}

//NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

//NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(rw http.ResponseWriter, r *http.Request) {
	remoteIp := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIp)

	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	renders.RenderTemplate(rw, r, "home.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) About(rw http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	remoteIp := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIp

	renders.RenderTemplate(rw, r, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) PostHome(rw http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		fmt.Fprint(rw, "Error parsing form")
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("firstname"),
		LastName:  r.Form.Get("lastname"),
		Email:     r.Form.Get("email"),
	}

	form := forms.New(r.PostForm)

	// form.Required("firstname", "lastname")
	// form.MinLength("firstname", 3, r)
	form.Has("firstname")
	form.Has("lastname")
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		renders.RenderTemplate(rw, r, "home.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	// m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(rw, r, "/reservation", http.StatusSeeOther)

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

	renders.RenderTemplate(rw, r, "reservation.page.html", &models.TemplateData{
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
		fmt.Println("could not pars json")
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(out)
}
