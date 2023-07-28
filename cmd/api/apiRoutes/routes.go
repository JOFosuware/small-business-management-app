package apiRoutes

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	apihandler "github.com/jofosuware/small-business-management-app/cmd/api/apiHandler"
)

func Routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Route("/api", func(mux chi.Router) {
		mux.Post("/customer-debt/{id}", apihandler.Repo.CustomerDebt)
	})
	return mux
}
