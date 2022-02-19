package main

import (
	"fmt"
	"testing"

	"github.com/amiranbari/bookings/pkg/config"
	"github.com/go-chi/chi/v5"
)

func TestRoute(t *testing.T) {
	var app config.AppConfig

	mux := route(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		//do nothing
	default:
		t.Error(fmt.Sprintf("type is not *chi.Mux:, but is %T", v))
	}
}
