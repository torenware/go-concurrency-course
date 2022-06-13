package main

import "net/http"

// Add session to the request
func (app *Config) AddSessionToRequest(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}
