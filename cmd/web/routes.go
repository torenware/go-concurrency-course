package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewMux()

	mux.Use(middleware.Recoverer)
	mux.Use(app.AddSessionToRequest)

	mux.Get("/", app.HomePage)

	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLogin)
	mux.Get("/logout", app.Logout)

	mux.Get("/register", app.Register)
	mux.Post("/register", app.PostRegister)

	mux.Get("/activate", app.ActivateUser)

	return mux
}
