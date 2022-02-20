package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestValid(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)

	form := New(r.PostForm)
	isValid := form.Valid()

	if !isValid {
		t.Error("Something error in form valid.")
	}
}

func TestRequired(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)

	form := New(r.PostForm)

	form.Required("a", "b", "c")

	if form.Valid() {
		t.Error("form shows valid when required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r, _ = http.NewRequest("POST", "whatever", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")

	if !form.Valid() {
		t.Error("shows does not have required fields when it does")
	}
}

func TestHas(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)

	form := New(r.PostForm)

	if form.Has("a") {
		t.Error("form shows valid when Has fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")

	r, _ = http.NewRequest("POST", "/whatever", nil)

	r.PostForm = postedData
	form = New(r.PostForm)

	form.Has("a")
	form.Has("b")

	if !form.Valid() {
		t.Error("shows does not have fields when it does")
	}

}

func TestMinLength(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)

	postedData := url.Values{}
	postedData.Add("a", "aa")
	r.PostForm = postedData
	form := New(r.PostForm)

	if !form.MinLength("a", 2) {
		t.Error("form has valid length when test say its not")
	}

	r, _ = http.NewRequest("POST", "/whatever", nil)

	postedData = url.Values{}
	postedData.Add("a", "a")
	r.PostForm = postedData
	form = New(r.PostForm)

	if form.MinLength("a", 2) {
		t.Error("form does not have valid length when test say it is")
	}
}

func TestIsEmail(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)

	form := New(r.PostForm)

	if form.IsEmail("email") {
		t.Error("form shows valid email when form does not have valid email")
	}

	r, _ = http.NewRequest("POST", "/whatever", nil)
	postedData := url.Values{}
	postedData.Add("email", "amiranbari33@gmail.com")
	r.PostForm = postedData
	form = New(r.PostForm)

	if !form.IsEmail("email") {
		t.Error("form shows valid email when form does not have valid email")
	}
}
