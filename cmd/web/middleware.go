package main

import "net/http"

// Add session to the request
func (app *Config) AddSessionToRequest(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}

// Enforce auth
func (app *Config) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "userID") {
			app.Session.Put(r.Context(), "error", "You must login to see that page.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
