package forms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGet(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)

	form := New(r.PostForm)

	fmt.Println(form.Errors.Get("a"))

	if form.Errors.Get("a") != "" {
		t.Error(" error when form does not any body")
	}

	r, _ = http.NewRequest("POST", "/whatever", nil)
	postedData := url.Values{}
	postedData.Add("a", "a")
	r.PostForm = postedData
	form = New(r.PostForm)

	form.MinLength("a", 2)

	if form.Errors.Get("a") == "" {
		t.Error("error does not have field when we sent")
	}
}
