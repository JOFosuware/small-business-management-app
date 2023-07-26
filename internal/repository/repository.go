package repository

import "github.com/jofosuware/small-business-management-app/internal/models"

type DatabaseRepo interface {
	AllUsers() bool

	InsertUser(u models.User) (int, error)
	Authenticate(username, password string) (int, string, error)
	FetchUser(username string) (models.User, error)
	InsertProduct(p models.Product) (models.Product, error)
	UpdateProduct(p models.Product) error
	FetchProduct(serial string) (models.Product, error)
	FetchAllProduct() ([]models.Product, error)
	DeleteProduct(serial string) error
	FetchCustomer(customerId string) (models.Customer, error)
	InsertCustomer(c models.Customer) (models.Customer, error)
	UpdateCustomer(c models.Customer) error
	FetchWitness(customerId string) (models.Witness, error)
	InsertWitnessData(w models.Witness) (models.Witness, error)
	UpdateWitness(w models.Witness) error
	InsertItem(itm models.Item) error
	UpdateItem(itm models.Item) error
	CustomerDebt(customerId string) ([]models.Item, error)
}
