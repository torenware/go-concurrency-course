package main

import (
	"context"
	"encoding/gob"
	"final-project/data"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
)

var testApp Config

func TestMain(m *testing.M) {

	gob.Register(data.User{})
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	// create waitgroup
	wg := sync.WaitGroup{}

	// set up the application config
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime|log.Ldate)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ltime|log.Ldate|log.Lshortfile)

	testApp = Config{
		Session:       session,
		Models:        data.NewTestModels(nil),
		Wait:          &wg,
		InfoLog:       infoLog,
		ErrorLog:      errorLog,
		ErrorChan:     make(chan error),
		ErrorChanDone: make(chan bool),
	}

	// error listener
	go func() {
		for {
			select {
			case err := <-testApp.ErrorChan:
				testApp.ErrorLog.Println(err)
			case <-testApp.ErrorChanDone:
				return
			}
		}
	}()

	// Mailer mock
	mailer := Mail{
		Wait:       &wg,
		MailerChan: make(chan Message),
		ErrorChan:  make(chan error),
		DoneChan:   make(chan bool),
	}
	testApp.Mailer = mailer

	// Mail listener
	go func() {
		for {
			select {
			case <-testApp.Mailer.MailerChan:
			case <-testApp.Mailer.ErrorChan:
			case <-testApp.Mailer.DoneChan:
				return
			}
		}
	}()

	os.Exit(m.Run())
}

// Create a Mock Context
func createMockContext(r *http.Request) context.Context {
	ctx, err := testApp.Session.Load(r.Context(), r.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
