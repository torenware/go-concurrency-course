package main

import "net/http"

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.gohtml", nil)
}

func (app *Config) PostLogin(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) Register(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegister(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) ActivateUser(w http.ResponseWriter, r *http.Request) {

}
