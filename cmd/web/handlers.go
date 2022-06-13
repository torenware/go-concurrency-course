package main

import (
	"database/sql"
	"final-project/data"
	"fmt"
	"net/http"
	"time"
)

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.gohtml", nil)
}

func (app *Config) PostLogin(w http.ResponseWriter, r *http.Request) {
	// renew token
	_ = app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println("problem parsing form:", err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		// set up a flash, redirect to login
		app.Session.Put(r.Context(), "error", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	matches, err := user.PasswordMatches(password)
	if err != nil {
		// set up flash etc.
		app.Session.Put(r.Context(), "error", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !matches {
		// for a test, let's assume we want to send a message every time
		// someone types in a wrong password. I'd hate to be on the
		// receiving end of this address :-)
		msg := Message{
			Subject: "Bad Login Attempt",
			To:      "faults@server-sec.com",
			Data:    fmt.Sprintf("Failed login by %s. Send the dogs.", user.Email),
		}
		app.sendMail(msg)

		app.Session.Put(r.Context(), "error", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return

	}

	app.Session.Put(r.Context(), "userID", user.ID)
	// user must be registered so the gob works. See main().
	app.Session.Put(r.Context(), "user", user)
	app.Session.Put(r.Context(), "flash", "Welcome User!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())
	app.Session.Put(r.Context(), "flash", "Goodbye!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) Register(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println("problem parsing form:", err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	verify_pw := r.Form.Get("verify-password")
	first := r.Form.Get("first-name")
	last := r.Form.Get("last-name")

	_, err = app.Models.User.GetByEmail(email)
	if err != sql.ErrNoRows {
		app.errorFlash(w, r, "Sorry! This email is not available", "/register")
		return
	}

	if password != verify_pw || password == "" {
		app.errorFlash(w, r, "Passwords required and must match", "/register")
		return
	}

	user := data.User{
		Email:     email,
		Password:  password,
		FirstName: first,
		LastName:  last,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	uid, err := app.Models.User.Insert(user)
	if err != nil {
		app.ErrorLog.Println("problem creating user:", err)
		app.errorFlash(w, r, "Sorry! Problem processing your registration", "/register")
		return
	}

	app.InfoLog.Printf("Mail would be sent for user %d", uid)

	url := fmt.Sprintf("http://localhost:8080/activate?email=%s", email)
	signedURL := GenerateTokenFromString(url)

	msg := Message{
		To:       email,
		Subject:  "Please verify your email",
		Template: "confirmation-email",
		Data:     signedURL,
	}

	app.sendMail(msg)
	app.Session.Put(r.Context(), "flash", "You would get a reg mail")

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *Config) ActivateUser(w http.ResponseWriter, r *http.Request) {
	url := r.RequestURI
	rebuiltURL := fmt.Sprintf("http://localhost:8080%s", url)
	okay := VerifyToken(rebuiltURL)

	if !okay {
		app.errorFlash(w, r, "Your confirmation link has expired or is invalid", "/")
		return
	}

	// Mark the user as activated and valid.
	email := r.URL.Query().Get("email")
	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.ErrorLog.Println("problem processing user", err)
		app.errorFlash(w, r, "Sorry! Problem handling your registration!", "/")
		return
	}

	user.Active = 1
	user.UpdatedAt = time.Now()

	err = user.Update()
	if err != nil {
		app.ErrorLog.Println("problem updating user", err)
		app.errorFlash(w, r, "Sorry! Problem handling your registration!", "/")
		return
	}
	msg := fmt.Sprintf("Welcome to the site, %s. You are now registered!", user.FirstName)
	app.Session.Put(r.Context(), "flash", msg)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
