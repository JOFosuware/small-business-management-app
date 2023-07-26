package models

import (
	"time"
)

// User Data struct
type User struct {
	FirstName string
	LastName  string
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Product Data struct
type Product struct {
	ID          int
	Serial      string
	Name        string
	Description string
	Price       float64
	Units       int32
	UserId      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Forms meta data struct
type FormMetaData struct {
	Message string
	Button  string
	Url     string
	Section string
}

// Client data struct
type Customer struct {
	ID           int
	CustomerId   string
	IDType       string
	FirstName    string
	LastName     string
	Phone        int
	HouseAddress string
	Location     string
	Landmark     string
	Agreement    string
	UserId       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Client Witness data struct
type Witness struct {
	ID         int
	CustomerId string
	FirstName  string
	LastName   string
	Phone      int
	Terms      string
	UserId     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Client's items credited data struct
type Item struct {
	ID         int       `json:"-"`
	CustomerId string    `json:"customerId"`
	Serial     string    `json:"serial"`
	Price      float32   `json:"price"`
	Quantity   int       `json:"quantity"`
	Total      float32   `json:"-"`
	Deposit    float32   `json:"deposit"`
	Balance    float32   `json:"balance"`
	UserId     int       `json:"-"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type Friends struct {
	Name  string `json:"name"`
	Place string `json:"place"`
	Year  string `json:"year"`
}
