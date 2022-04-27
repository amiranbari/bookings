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
	"time"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{"home", "/", http.StatusOK},
	{"about", "/about", http.StatusOK},
	{"json", "/json", http.StatusOK},
	{"reservation", "/reservation", http.StatusOK},
	{"non-existent", "/dark/mode", http.StatusNotFound},
	{"login", "/login", http.StatusOK},
	{"logout", "/logout", http.StatusOK},
	{"dashboard", "/admin/dashboard", http.StatusOK},
	{"admin-reservations", "/admin/reservations", http.StatusOK},
	{"admin-new-reservations", "/admin/new-reservations", http.StatusOK},
	{"admin-show-reservation", "/admin/reservations/1", http.StatusOK},
	{"admin-fail-reservation", "/admin/reservations/non-roomID", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()

	ts := httptest.NewTLSServer(routes)

	defer ts.Close()

	for _, e := range theTests {
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
}

func TestAdminPostShowReservations(t *testing.T) {
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

func TestAdminPutShowReservations(t *testing.T) {
	req, _ := http.NewRequest("GET", "/admin/reservations/1/processed", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/admin/reservations/1/processed"

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.AdminPutShowReservations)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("AdminPutShowReservations Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//invalid-reservation ID
	req, _ = http.NewRequest("GET", "/admin/reservations/invalid-ID/processed", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/admin/reservations/invalid-ID/processed"

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("AdminPutShowReservations Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//invalid-reservation ID in database
	req, _ = http.NewRequest("GET", "/admin/reservations/2/processed", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/admin/reservations/2/processed"

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("AdminPutShowReservations Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestAdminDeleteReservation(t *testing.T) {
	req, _ := http.NewRequest("GET", "/admin/reservations/1/delete", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/admin/reservations/1/delete"

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.AdminDeleteReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("AdminDeleteReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//invalid-reservation ID
	req, _ = http.NewRequest("GET", "/admin/reservations/invalid-ID/delete", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/admin/reservations/invalid-ID/delete"

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("AdminDeleteReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//invalid-reservation ID in database
	req, _ = http.NewRequest("GET", "/admin/reservations/2/delete", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/admin/reservations/2/delete"

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("AdminDeleteReservation Handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

var adminPostReservationCalendarTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
	blocks               int
	reservations         int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
	},
	{
		name:                 "cal-blocks",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		blocks:               1,
	},
	{
		name:                 "cal-res",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		reservations:         1,
	},
}

func TestPostReservationCalendar(t *testing.T) {
	for _, e := range adminPostReservationCalendarTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calender", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calender", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalender)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
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
