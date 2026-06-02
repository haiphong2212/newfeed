package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFRequiresHeaderForCookieAuthenticatedStateChange(t *testing.T) {
	handler := CSRF(CSRFOptions{
		CookieName:      "csrf_token",
		AuthCookieNames: []string{"newfeed_access_token"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "newfeed_access_token", Value: "access"})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without csrf header, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-CSRF-Token", "token")
	req.AddCookie(&http.Cookie{Name: "newfeed_access_token", Value: "access"})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token"})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 with csrf header, got %d", rec.Code)
	}
}

func TestCSRFSkipsUnauthenticatedRequests(t *testing.T) {
	handler := CSRF(CSRFOptions{
		CookieName:      "csrf_token",
		AuthCookieNames: []string{"newfeed_access_token"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected csrf skip for no auth cookie, got %d", rec.Code)
	}
}
