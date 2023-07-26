package middleware

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/jofosuware/small-business-management-app/internal/helpers"
	"github.com/justinas/nosurf"
)

var App config.AppConfig
var Session *scs.SessionManager

// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   App.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return Session.LoadAndSave(next)
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			Session.Put(r.Context(), "error", "Log in first")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
