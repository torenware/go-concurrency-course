package main

import (
	"net/http"
	"time"

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

	mux.Get("/plans", app.ChoosePlans)
	mux.Get("/subscribe", app.SubscribePlan)

	// test for mail
	mux.Get("/test-email", func(w http.ResponseWriter, r *http.Request) {
		m := Mail{
			Domain:      "localhost",
			Host:        "localhost",
			Port:        1025,
			FromAddress: "joe@mamma.org",
			FromName:    "Joe Yo",
			Encryption:  "none",
			ErrorChan:   make(chan error),
		}

		msg := Message{
			To:      "killroy@here.was",
			Data:    "Where you was?",
			Subject: "Yer location",
		}
		errChan := make(chan error)
		go m.sendMail(msg, errChan)

		select {
		case err := <-errChan:
			app.ErrorLog.Println("Mail got error:", err)
		case <-time.After(5 * time.Second):
			app.InfoLog.Println("sendMail timed out")

		}

		app.InfoLog.Println("We tried to send mail")
	})

	return mux
}
