package main

import (
	"errors"
	"final-project/data"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

// Directories changable for testing:
var tempDirectory = "./tmp"
var pdfDirectory = "./pdfs"

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

	matches, err := app.Models.User.PasswordMatches(*user, password)
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
	app.Session.Put(r.Context(), "user", *user)
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
	if err != nil {
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
	NewURLSigner()
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
	NewURLSigner()
	okay := VerifyToken(rebuiltURL)

	if !okay {
		app.errorFlash(w, r, "Your confirmation link has expired or is invalid", "/")
		return
	}

	if Expired(rebuiltURL, 60) {
		app.errorFlash(w, r, "Your confirmation link has expired!", "/")
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

	if user.Active == 1 {
		app.Session.Put(r.Context(), "flash", "You are already registered!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user.Active = 1
	user.UpdatedAt = time.Now()

	err = app.Models.User.Update(*user)
	if err != nil {
		app.ErrorLog.Println("problem updating user", err)
		app.errorFlash(w, r, "Sorry! Problem handling your registration!", "/")
		return
	}
	msg := fmt.Sprintf("Welcome to the site, %s. You are now registered!", user.FirstName)
	app.Session.Put(r.Context(), "flash", msg)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) ChoosePlans(w http.ResponseWriter, r *http.Request) {

	plans, err := app.Models.Plan.GetAll()
	if err != nil {
		app.errorFlash(w, r, "Sorry! Could not display this page", "/")
		return
	}

	data := map[string]any{
		"Plans": plans,
	}

	app.render(w, r, "plans.page.gohtml", &TemplateData{
		Data: data,
	})
}

func (app *Config) SubscribePlan(w http.ResponseWriter, r *http.Request) {

	planParam := r.URL.Query().Get("plan")
	planID, err := strconv.Atoi(planParam)
	if err != nil {
		app.ErrorLog.Println("subscribe passed wrong parameter")
		app.errorFlash(w, r, "Cannot subscribe to that plan.", "/members/plans")
		return
	}

	plan, err := app.Models.Plan.GetOne(planID)
	if err != nil {
		app.ErrorLog.Printf("subscribe passed unavailable plan %d with error %v", planID, err)
		app.errorFlash(w, r, "Cannot subscribe to that plan.", "/members/plans")
		return
	}

	user, ok := app.Session.Get(r.Context(), "user").(data.User)
	if !ok {
		app.ErrorLog.Println("user not in session?")
		app.errorFlash(w, r, "Please log in.", "/login")
		return
	}

	err = app.Models.Plan.SubscribeUserToPlan(user, *plan)
	if err != nil {
		app.ErrorLog.Printf("could not subscribe: %v", err)
		app.errorFlash(w, r, "Cannot subscribe to that plan.", "/members/plans")
		return
	}

	// Generate an invoice
	app.Wait.Add(1)
	go func() {
		defer app.Wait.Done()

		invoice, err := app.GenerateInvoice(user, *plan)
		if err != nil {
			app.ErrorChan <- err
		}

		// send an email
		msg := Message{
			To:       user.Email,
			Subject:  fmt.Sprintf("You've Subscribed to Our %s", plan.PlanName),
			Data:     invoice,
			Template: "invoice",
		}
		// kick the invoice off to its own routine.
		app.sendMail(msg)
	}()

	// generate a customized manual PDF
	app.Wait.Add(1)
	go func() {
		defer app.Wait.Done()

		pdf := app.GenerateManual(user, plan)
		tmpFile := fmt.Sprintf("%s/%d_user-manual.pdf", tempDirectory, user.ID)
		err := pdf.OutputFileAndClose(tmpFile)
		if err != nil {
			app.ErrorChan <- err
			return
		}

		msg := Message{
			To:      user.Email,
			Subject: fmt.Sprintf("Your %s User Manual", plan.PlanName),
			Data:    "Your personalized manual is attached:",
			AttachmentMap: map[string]string{
				"Manual.pdf": tmpFile,
			},
		}

		app.sendMail(msg)

		// temp: an error test
		app.ErrorChan <- errors.New("a custom error test")
	}()

	// update the user in session, since it has updated.
	userPtr, err := app.Models.User.GetOne(user.ID)
	if err != nil {
		// this is a convenience, so if there's an error,
		// log it and ignore.
		app.ErrorLog.Printf("error retrieving updated user: %v", err)
	} else {
		app.Session.Put(r.Context(), "user", *userPtr)
	}

	app.Session.Put(r.Context(), "flash", fmt.Sprintf("You are subscribed to %s", plan.PlanName))
	http.Redirect(w, r, "/members/plans", http.StatusSeeOther)
}

func (app *Config) GenerateInvoice(user data.User, plan data.Plan) (string, error) {
	// We punt!
	return plan.PlanAmountFormatted, nil
}

func (app *Config) GenerateManual(user data.User, plan *data.Plan) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10) //in mm as specified

	importer := gofpdi.NewImporter()

	// simulate a complex PDF...
	time.Sleep(5 * time.Second)

	t := importer.ImportPage(pdf, fmt.Sprintf("%s/manual.pdf", pdfDirectory), 1, "/MediaBox")
	pdf.AddPage()

	// center where we are writing
	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

	// Put down the pen
	pdf.SetX(75)
	pdf.SetY(150)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s %s", user.FirstName, user.LastName), "", "C", false)
	pdf.Ln(5)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s User Guide", plan.PlanName), "", "C", false)

	return pdf
}
