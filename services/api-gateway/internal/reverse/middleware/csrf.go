package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/response"
)

type CSRFOptions struct {
	CookieName      string
	AuthCookieNames []string
	CookieSecure    bool
}

func CSRF(opts CSRFOptions) Middleware {
	if opts.CookieName == "" {
		opts.CookieName = "csrf_token"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !stateChanging(r.Method) || !hasAuthCookie(r, opts.AuthCookieNames) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(opts.CookieName)
			if err != nil || cookie.Value == "" || r.Header.Get("X-CSRF-Token") != cookie.Value {
				response.Error(w, http.StatusForbidden, "invalid csrf token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func TokenHandler(cookieName string, secure bool) http.HandlerFunc {
	if cookieName == "" {
		cookieName = "csrf_token"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		token := csrfToken()
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			MaxAge:   int((2 * time.Hour).Seconds()),
			Secure:   secure,
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		})
		response.JSON(w, http.StatusOK, map[string]string{"csrf_token": token})
	}
}

func stateChanging(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func hasAuthCookie(r *http.Request, names []string) bool {
	for _, name := range names {
		if cookie, err := r.Cookie(name); err == nil && cookie.Value != "" {
			return true
		}
	}
	return false
}

func csrfToken() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}
