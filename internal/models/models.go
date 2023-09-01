package models

import (
	"time"
)

// User Data struct
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Username    string
	Password    string
	AccessLevel string
	Image       []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

type PageTitle struct {
	Main        string
	Sub         string
	Description string
	PlaceHolder string
}

// Client data struct
type Customer struct {
	ID              int
	CustomerId      string
	CustImage       []byte
	IDType          string
	CardImage       []byte
	FirstName       string
	LastName        string
	Phone           int
	HouseAddress    string
	Location        string
	Landmark        string
	Status          string
	Agreement       string
	UserId          int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CustImgString   string
	CardImgString   string
	CreatedAtString string
	UpdatedAtString string
}

// Client Witness data struct
type Witness struct {
	ID              int
	CustomerId      string
	FirstName       string
	LastName        string
	Phone           int
	Terms           string
	Image           []byte
	UserId          int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ImageString     string
	CreatedAtString string
	UpdatedAtString string
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
	Image      []byte    `json:"-"`
	UserId     int       `json:"-"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type Friends struct {
	Name  string `json:"name"`
	Place string `json:"place"`
	Year  string `json:"year"`
}

// Payment is the model type for payment database
type Payments struct {
	CustomerId string
	Month      string
	Amount     int
	Date       time.Time
	UserId     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Purchases struct {
	Serial    string
	Quantity  int
	Amount    int
	UserId    int
	CreatedAt time.Time
	UpdatedAt time.Time
}
