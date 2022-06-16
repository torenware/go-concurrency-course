package main

import (
	"final-project/data"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Get pages
var pages = []struct {
	Page          string
	URL           string
	Handler       http.HandlerFunc
	ExpectedCode  int
	SessionBefore map[string]any
	// existing or non-existing keys.
	// non-existing start with "!"
	SessionAfter []string
	ExpectedHTML string
}{
	{
		Page:         "home",
		URL:          "/",
		Handler:      testApp.HomePage,
		ExpectedCode: http.StatusOK,
	},
	{
		Page:         "login",
		URL:          "/login",
		Handler:      testApp.LoginPage,
		ExpectedCode: http.StatusOK,
	},
	{
		Page:         "register",
		URL:          "/register",
		Handler:      testApp.Register,
		ExpectedCode: http.StatusOK,
		ExpectedHTML: `>Register</h1>`,
	},
	{
		Page:         "logout",
		URL:          "/logout",
		Handler:      testApp.Logout,
		ExpectedCode: http.StatusSeeOther,
		SessionBefore: map[string]any{
			"userID": 1,
			"user":   data.User{},
		},
		SessionAfter: []string{
			"!userID",
			"!user",
		},
	},
}

func TestHandlers_GetPages(t *testing.T) {
	pathToTemplates = "./templates"

	for _, page := range pages {
		// t.Log("Testing page =", page.Page)
		req, _ := http.NewRequest("GET", page.URL, nil)
		ctx := createMockContext(req)
		req = req.WithContext(ctx)

		if len(page.SessionBefore) > 0 {
			for key, val := range page.SessionBefore {
				testApp.Session.Put(ctx, key, val)
			}
		}

		// and we need a writer
		rr := httptest.NewRecorder()

		page.Handler.ServeHTTP(rr, req)

		if len(page.SessionAfter) > 0 {
			for _, key := range page.SessionAfter {
				if string(key[0]) == "!" {
					key = key[1:]
					if testApp.Session.Exists(ctx, key) {
						t.Errorf("%s: expected key '%s' to be gone, but it remains", page.Page, key)
					}
				} else if !testApp.Session.Exists(ctx, key) {
					t.Errorf("%s: expected key '%s' to be present", page.Page, key)
				}
			}
		}

		resp := rr.Result()

		if len(page.ExpectedHTML) > 0 {
			body, _ := io.ReadAll(resp.Body)
			if !strings.Contains(string(body), page.ExpectedHTML) {
				t.Errorf("%s: expected body to contain string '%s'", page.Page, page.ExpectedHTML)
			}
		}

		if rr.Code != page.ExpectedCode {
			t.Errorf("%s: expected %d, got %d", page.Page, page.ExpectedCode, rr.Code)
		}
	}

}

func TestHandlers_PostLogin(t *testing.T) {
	formPost := url.Values{}
	formPost.Add("email", "who@first.com")
	formPost.Add("password", "it-is-a-secret")

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(formPost.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	ctx := createMockContext(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.PostLogin)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("%s: expected %d but got %d", "post-login", http.StatusSeeOther, rr.Code)
	}

	// and we should have a flash
	if !testApp.Session.Exists(ctx, "flash") {
		t.Errorf("%s: login was not successful", "post-login")
	}

	if !testApp.Session.Exists(ctx, "userID") {
		t.Errorf("%s: session still lacks a userID now", "post-login")
	}

	if !testApp.Session.Exists(ctx, "user") {
		t.Errorf("%s: session still lacks a user object now", "post-login")
	}

	_, ok := testApp.Session.Get(ctx, "user").(data.User)
	if !ok {
		t.Errorf("%s: user in session is not a user object", "post-login")
	}

}

func TestHandlers_ChoosePlans(t *testing.T) {
	pathToTemplates = "./templates"

	req, _ := http.NewRequest("GET", "/members/plans", nil)
	ctx := createMockContext(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(testApp.ChoosePlans)
	// wrap the handler in the middleware
	wrapped := testApp.Auth(nextHandler)
	wrapped.ServeHTTP(rr, req)

	// We should not be allowed to this page w/o credentials
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("choose-plans: expected redirect since not logged in, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	testApp.Session.Put(ctx, "userID", 1)
	testApp.Session.Put(ctx, "user", data.User{})
	req = req.WithContext(ctx)
	wrapped.ServeHTTP(rr, req)

	// We should now be allowed to this page
	if rr.Code != http.StatusOK {
		t.Errorf("choose-plans: expected to see this page, but got %d", rr.Code)
	}

}

func TestHandlers_SubscribePlan(t *testing.T) {
	mailMessages = []Message{}

	req, _ := http.NewRequest("GET", "/members/subscribe?plan=3", nil)
	ctx := createMockContext(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	testApp.Session.Put(ctx, "userID", 1)
	testApp.Session.Put(ctx, "user", data.User{
		ID:        99,
		FirstName: "Frederick",
		LastName:  "Fronkenstein",
	})

	handler := http.HandlerFunc(testApp.SubscribePlan)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("subscribe-plan: expected 'see other', got %d", rr.Code)
	}

	if !testApp.Session.Exists(ctx, "flash") {
		t.Error("subscribe-plan: expected a flash message on success")
	}

	// Make sure the WG clears...
	wgDone := make(chan bool)

	go func() {
		testApp.Wait.Wait()
		wgDone <- true
	}()

	select {
	case <-wgDone:
	case <-time.After(10 * time.Second):
		t.Error("subscribe-plan: waitgroup did not release; timing out.")
	}
	testApp.InfoLog.Println("wait group released.")

	if len(mailMessages) != 2 {
		t.Errorf("subscribe-plan: expected 2 mail messages, got %d", len(mailMessages))
	}

}

func TestHandlers_PostRegister(t *testing.T) {
	mailMessages = []Message{}

	formPost := url.Values{}
	formPost.Add("email", "who@first.com")
	formPost.Add("password", "it-is-a-secret")
	formPost.Add("verify-password", "it-is-a-secret")
	formPost.Add("first-name", "Lois")
	formPost.Add("last-name", "Lane")

	req, _ := http.NewRequest("POST", "/register", strings.NewReader(formPost.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	ctx := createMockContext(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.PostRegister)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("post-register: expected redirect, got %d", rr.Code)
	}

	resp := rr.Result()
	location := resp.Header.Get("Location")
	if location != "/" {
		t.Errorf("post-register: expected redirect to /, got %s", location)
	}

	if !testApp.Session.Exists(ctx, "flash") {
		t.Error("post-register: did not get success message")
	}

	// Make sure the WG clears...
	wgDone := make(chan bool)

	go func() {
		testApp.Wait.Wait()
		wgDone <- true
	}()

	select {
	case <-wgDone:
	case <-time.After(10 * time.Second):
		t.Error("post-register: waitgroup did not release; timing out.")
	}
	testApp.InfoLog.Println("wait group released.")

	if len(mailMessages) != 1 {
		t.Errorf("post-register: expected 1 mail message, got %d", len(mailMessages))
	}

	data, ok := mailMessages[0].Data.(string)

	if !ok {
		t.Error("post-register: could not get data back from mail message")
	}

	if !VerifyToken(data) {
		t.Error("post-register: did not get signed URL from message")
	}

}
