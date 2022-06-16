package main

import (
	"final-project/data"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
