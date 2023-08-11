package repository

import "github.com/jofosuware/small-business-management-app/internal/models"

type DatabaseRepo interface {
	AllUsers() bool

	InsertUser(u models.User) (int, error)
	Authenticate(username, password string) (models.User, error)
	FetchUser(username string) (models.User, error)
	InsertProduct(p models.Product) (models.Product, error)
	UpdateProduct(models.Product) error
	DecreaseQuantity(models.Product) error
	IncreaseQuantity(models.Product) error
	FetchProduct(serial string) (models.Product, error)
	FetchAllProduct() ([]models.Product, error)
	DeleteProduct(serial string) error
	FetchCustomer(customerId string) (models.Customer, error)
	InsertCustomer(c models.Customer) error
	UpdateCustomer(c models.Customer) error
	UpdateContactStatus(models.Customer) error
	FetchWitness(customerId string) (models.Witness, error)
	InsertWitnessData(w models.Witness) (models.Witness, error)
	UpdateWitness(w models.Witness) error
	InsertItem(itm models.Item) error
	UpdateItem(itm models.Item) error
	UpdateBalance(itm models.Item) error
	CustomerDebt(customerId string) ([]models.Item, error)
	InsertPayment(p models.Payments) error
	CustomerPayment(customerId string) ([]models.Payments, error)
	InsertPurchase(models.Purchases) (int, error)
	DeletePurchase(int) error
}
