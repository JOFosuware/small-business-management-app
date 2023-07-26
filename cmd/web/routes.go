package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/jofosuware/small-business-management-app/internal/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Post("/login", handlers.Repo.PostLogin)
	mux.Get("/logout", handlers.Repo.Logout)
	mux.Post("/signup", handlers.Repo.AddUser)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.Dashboard)

		//Product route
		mux.Get("/add-product", handlers.Repo.ProductForm)
		mux.Post("/add-product", handlers.Repo.AddProduct)
		mux.Get("/search-product", handlers.Repo.SearchProduct)
		mux.Post("/edit-form", handlers.Repo.EditProductForm)
		mux.Get("/edit-product", handlers.Repo.ProductForm)
		mux.Post("/edit-product", handlers.Repo.UpdateProduct)
		mux.Get("/delete-product", handlers.Repo.SearchProduct)
		mux.Post("/delete-product", handlers.Repo.DeleteProduct)

		//Contract Route
		mux.Get("/add-contract", handlers.Repo.CustomerForm)
		mux.Post("/add-contract", handlers.Repo.PostCustomer)
		mux.Get("/edit-contract", handlers.Repo.CustomerForm)
		mux.Post("/update-contract", handlers.Repo.UpdateCustomer)
		mux.Get("/add-witness", handlers.Repo.GetWitnessForm)
		mux.Post("/add-witness", handlers.Repo.PostWitness)
		mux.Get("/edit-witness", handlers.Repo.GetWitnessForm)
		mux.Post("/update-witness", handlers.Repo.UpdateWitness)
		mux.Get("/add-item", handlers.Repo.ItemForm)
		mux.Post("/add-item", handlers.Repo.PostItem)
		mux.Get("/edit-item", handlers.Repo.ItemForm)
		mux.Post("/edit-item", handlers.Repo.UpdateItem)
		mux.Get("/pay", handlers.Repo.PaymentForm)
		mux.Post("/customer-debt/{id}", handlers.Repo.CustomerDebt)
	})

	return mux
}
