package main

import "net/http"

// wrap our mailer so that we don't forget to
// add to the WaitGroup; mailer.sendMail() decrements.
func (app *Config) sendMail(msg Message) {
	app.Mailer.Wait.Add(1)
	app.Mailer.MailerChan <- msg
}

func (app *Config) errorFlash(w http.ResponseWriter, r *http.Request, msg, url string) {
	app.Session.Put(r.Context(), "error", msg)
	http.Redirect(w, r, url, http.StatusSeeOther)
}
