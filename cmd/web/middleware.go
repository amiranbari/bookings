package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func NoSruve(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

//laods and saves the session on every reqeust
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
} 
