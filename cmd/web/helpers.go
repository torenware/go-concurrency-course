package main

// wrap our mailer so that we don't forget to
// add to the WaitGroup; mailer.sendMail() decrements.
func (app *Config) sendMail(msg Message) {
	app.Mailer.Wait.Add(1)
	app.Mailer.MailerChan <- msg
}
