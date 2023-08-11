package routes

import (
	"net/http"

	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/jofosuware/small-business-management-app/cmd/web/middleware"
	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/jofosuware/small-business-management-app/internal/handlers"
)

func Routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(chiMiddleware.Recoverer)
	//mux.Use(middleware.NoSurf)
	mux.Use(middleware.SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Post("/login", handlers.Repo.PostLogin)
	mux.Get("/logout", handlers.Repo.Logout)
	mux.Get("/signup", handlers.Repo.UserForm)
	mux.Get("/edit-user", handlers.Repo.UserForm)
	mux.Post("/signup", handlers.Repo.PostUser)
	mux.Post("/developer", handlers.Repo.PostDeveloper)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(middleware.Auth)
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
		mux.Get("/increase-qty", handlers.Repo.IncreaseQtyForm)
		mux.Post("/increase-qty", handlers.Repo.PostIncreaseQty)

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
		mux.Post("/pay", handlers.Repo.PostPayments)
		mux.Get("/generate-receipt", handlers.Repo.ReceiptPage)

		//Purchase Route
		mux.Get("/add-purchase", handlers.Repo.PurchaseForm)
		mux.Post("/add-purchase", handlers.Repo.PostPurchase)
	})

	return mux
}
