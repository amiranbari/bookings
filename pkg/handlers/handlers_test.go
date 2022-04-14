package handlers

import (
	"context"
	"fmt"
	"github.com/amiranbari/bookings/pkg/models"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name   string
	url    string
	method string
	//params 			   []string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"json", "/json", "GET", http.StatusOK},
	{"reservation", "/reservation", "GET", http.StatusOK},
	{"non-existent", "/dark/mode", "GET", http.StatusNotFound},
	{"login", "/login", "GET", http.StatusOK},
	{"logout", "/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"admin-reservations", "/admin/reservations", "GET", http.StatusOK},
	{"admin-new-reservations", "/admin/new-reservations", "GET", http.StatusOK},
	{"admin-show-reservation", "/admin/reservations/1", "GET", http.StatusOK},
	{"admin-fail-reservation", "/admin/reservations/non-roomID", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()

	ts := httptest.NewTLSServer(routes)

	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}

			//} else {
			//
			//	values := url.Values{}
			//	for _, x := range e.params {
			//		values.Add(x.key, x.value)
			//	}
			//
			//	resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			//
			//	if err != nil {
			//		t.Log(err)
			//		t.Fatal(err)
			//	}
			//
			//	if resp.StatusCode != e.expectedStatusCode {
			//		t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			//	}
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:    1,
			Title: "General",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.MakeReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Reservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test with none existing room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomId = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test with existing room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomId = 1
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Reservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

}

func TestRepository_PostReservation(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-01")
	postedData.Add("firstname", "amir")
	postedData.Add("lastname", "anbari")
	postedData.Add("email", "amir@gmail.com")
	postedData.Add("phone", "+989335716724")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:    1,
			Title: "General",
		},
	}
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test for missing body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//set just firstname - !valid form
	reqBody := "firstname=amir"

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test missing session
	reqBody = "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-01")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "firstname=amir")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "lastname=anbari")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=amir@gmail.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=+989335716724")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//error in inserting reservation
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	reservation.RoomId = 2
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//error in inserting restriction
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	reservation.RoomId = 100
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostSearch(t *testing.T) {
	reqBody := "start_date=2040-02-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2040-02-01")

	req, _ := http.NewRequest("POST", "/search", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostSearch)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("PostSearch Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	//test missing body
	req, _ = http.NewRequest("POST", "/search", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostSearch)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostSearch Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test parameters not valid
	reqBody = "room=2020-01-01"
	req, _ = http.NewRequest("POST", "/search", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostSearch)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostSearch Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test error in search availability room
	reqBody = "start_date=2040-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2040-01-01")
	req, _ = http.NewRequest("POST", "/search", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostSearch)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostSearch Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//test empty searching room
	reqBody = "start_date=2021-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2021-01-01")
	req, _ = http.NewRequest("POST", "/search", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostSearch)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostSearch Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

}

func TestRepository_Search(t *testing.T) {

	req, _ := http.NewRequest("POST", "/search", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.Search)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Search Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"

	rr := httptest.NewRecorder()

	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:    1,
			Title: "General",
		},
	}
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//testing with wrong id
	req, _ = http.NewRequest("GET", "/choose-room/!valid", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/!valid"

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//testing with not set session
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHtml       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"admin@gmail.com",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"me@gmail.com",
		http.StatusSeeOther,
		"",
		"/login",
	},
	{
		"invalid-data",
		"j",
		http.StatusOK,
		"",
		"",
	},
}

func TestLogin(t *testing.T) {
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "12345678")

		//request
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		//header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		//call the handler
		handler := http.HandlerFunc(Repo.PostLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("Faild %s, Post Login Handler return wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("Failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHtml != "" {
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHtml) {
				t.Errorf("Failed %s: expected html %s, but got %s", e.name, e.expectedHtml, html)
			}
		}
	}
}

var adminPostShowReservationTests = []struct {
	url                string
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedHtml       string
	expectedLocation   string
}{
	{
		"/admin/reservations/1",
		"valid data",
		url.Values{
			"first_name": {"Amir"},
			"last_name":  {"Anbari"},
			"email":      {"amir@anbari.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusSeeOther,
		"",
		"/admin/reservations",
	},
	{
		"/admin/reservations/2",
		"invalid-roomID",
		url.Values{
			"first_name": {"Amir"},
			"last_name":  {"Anbari"},
			"email":      {"amir@anbari.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusSeeOther,
		"",
		"/admin/reservations",
	},
	//{
	//	"invalid-data",
	//	url.Values{
	//		"first_name": {"Amir"},
	//		"last_name":  {"Anbari"},
	//		"email":      {"amir@anbari.com"},
	//		"phone":      {"555-555-5555"},
	//	},
	//	http.StatusSeeOther,
	//	"",
	//	"",
	//},
}

func TestRepository_AdminPostShowReservations(t *testing.T) {
	for _, e := range adminPostShowReservationTests {
		//request
		req, _ := http.NewRequest("POST", e.url, strings.NewReader(e.postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = e.url

		//header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		//call the handler
		handler := http.HandlerFunc(Repo.AdminPostShowReservations)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("Faild %s, Post Login Handler return wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("Failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHtml != "" {
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHtml) {
				t.Errorf("Failed %s: expected html %s, but got %s", e.name, e.expectedHtml, html)
			}
		}
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
