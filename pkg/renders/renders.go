package renders

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/amiranbari/bookings/pkg/config"
	"github.com/amiranbari/bookings/pkg/models"
	"github.com/justinas/nosurf"
)

var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
}

var app *config.AppConfig

func Iterate(count int) []int {
	var i int
	var items []int
	for i = 1; i <= count; i++ {
		items = append(items, i)
	}
	return items
}

func HumanDate(t time.Time) string {
	return t.Format("2006-02-01")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func NewRenderer(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {

	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")

	td.CSRFToken = nosurf.Token(r)

	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = true
	}

	return td
}

func Template(rw http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {

	var tc config.TemplateCache

	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]

	if !ok {
		return errors.New("Could not get template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(rw)

	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return nil
	}

	return nil
}

func CreateTemplateCache() (config.TemplateCache, error) {
	myCache := config.TemplateCache{}

	pages, err := filepath.Glob("../../templates/*.page.html")

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob("../../templates/*.layout.html")

		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("../../templates/*.layout.html")

			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
