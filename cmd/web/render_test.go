package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfig_AddDefaultData(t *testing.T) {
	// URL is arbitrary and no body, so no reader.
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := createMockContext(req)
	req = req.WithContext(ctx)

	testApp.Session.Put(ctx, "flash", "flash")
	testApp.Session.Put(ctx, "warning", "warning")
	testApp.Session.Put(ctx, "error", "error")

	td := testApp.AddDefaultData(&TemplateData{}, req)

	if td.Flash != "flash" {
		t.Error("Flash did not come across into template data")
	}

	if td.Warning != "warning" {
		t.Error("Warning did not come across into template data")
	}

	if td.Error != "error" {
		t.Error("Error did not come across into template data")
	}

}

func TestConfig_IsAuthenticated(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := createMockContext(req)
	req = req.WithContext(ctx)

	// Empty session, so should not auth
	auth := testApp.IsAuthenticated(req)
	if auth {
		t.Error("expecting not authed, but authed")
	}

	// Make it look authed.
	testApp.Session.Put(ctx, "userID", 1)
	auth = testApp.IsAuthenticated(req)
	if !auth {
		t.Error("expecting authed, but not authed")
	}
}

func TestConfig_render(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := createMockContext(req)
	req = req.WithContext(ctx)

	// and we need a writer
	rr := httptest.NewRecorder()

	// adjust location of template directory
	pathToTemplates = "./templates"

	// page does not exist
	testApp.render(rr, req, "fake.page.gohtml", nil)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(
			"Expected internal server error %d, got %d",
			http.StatusInternalServerError,
			rr.Code,
		)
	}
	// legit page
	rr = httptest.NewRecorder()
	testApp.render(rr, req, "home.page.gohtml", nil)

	if rr.Code != http.StatusOK {
		t.Errorf(
			"Expected OK %d, got %d",
			http.StatusOK,
			rr.Code,
		)
	}

}
