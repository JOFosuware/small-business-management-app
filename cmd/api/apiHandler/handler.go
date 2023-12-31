package apihandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/jofosuware/small-business-management-app/internal/helpers"
	"github.com/jofosuware/small-business-management-app/internal/models"
	"github.com/jofosuware/small-business-management-app/internal/render"
	"github.com/jofosuware/small-business-management-app/internal/repository"
)

var Repo Repository

type Repository struct {
	DB       repository.DatabaseRepo
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

// CustomerDebt handles the request for the customer balance
func (c *Repository) CustomerDebt(w http.ResponseWriter, r *http.Request) {
	custId := chi.URLParam(r, "id")

	type payload struct {
		Err     bool    `json:"error"`
		Message string  `json:"message"`
		Debt    float64 `json:"debt,omitempty"`
		Payment float64 `json:"payment,omitempty"`
	}

	balance, err := c.CalcCustomerDebt(custId)
	if err != nil {
		payload := payload{
			Err:     true,
			Message: fmt.Sprintf("%s", err),
		}
		c.ErrorLog.Println(err)
		jsonData, _ := json.Marshal(payload)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	cust, err := c.DB.FetchCustomer(custId)
	if err != nil {
		payload := payload{
			Err:     true,
			Message: fmt.Sprintf("%s", err),
		}
		c.ErrorLog.Println(err)
		jsonData, _ := json.Marshal(payload)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if cust.Months == 0 {
		payload := payload{
			Err:     true,
			Message: "Customer is fully paid",
		}
		jsonData, _ := json.Marshal(payload)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	pload := payload{
		Err:     false,
		Message: "",
		Debt:    helpers.ToDecimalPlace(balance, 2),
		Payment: helpers.ToDecimalPlace(balance/float64(cust.Months), 2),
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// CalcCustomerDebt calculates customer debt to the enterprise
func (c *Repository) CalcCustomerDebt(customerId string) (float64, error) {
	if customerId == "" {
		return 0, errors.New("customer ID not provided")
	}

	custDebt, err := c.DB.CustomerDebt(customerId)
	if err != nil {
		return 0, errors.New("customer debt information can't be retrieved")
	}

	if len(custDebt) == 0 {
		return 0, fmt.Errorf("customer with this id: %s does not owe", customerId)
	}

	custPymt, err := c.DB.CustomerPayment(customerId)
	if err != nil {
		return 0, errors.New("customer payments information can't be retrieved")
	}

	amount := 0.00
	if len(custPymt) != 0 {
		for _, v := range custPymt {
			amount += float64(v.Amount)
		}
	}

	balance := 0.00
	for _, v := range custDebt {
		balance += float64(v.Balance)
	}

	balance -= amount
	return balance, nil
}

// CustomerOwingToday handles the request for the customers owing at the present day
func (c *Repository) CustomerOwingToday(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		Err      bool            `json:"error"`
		Message  string          `json:"message"`
		Debt     float64         `json:"debt,omitempty"`
		Payment  float64         `json:"payment"`
		Customer models.Customer `json:"customer,omitempty"`
	}

	var pload []payload
	custs, err := c.DB.FetchAllCustomers()
	if err != nil {
		payload := payload{
			Err:     true,
			Message: fmt.Sprintf("%s", err),
		}

		jsonData, _ := json.Marshal(payload)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	for _, v := range custs {
		presentDay := time.Now().Day()
		pastDay := v.CreatedAt.Day()

		if presentDay == pastDay && v.Status == "on_contract" {
			balance, _ := c.CalcCustomerDebt(v.CustomerId)

			pload = append(pload, payload{
				Err:      false,
				Message:  "",
				Debt:     helpers.ToDecimalPlace(balance, 2),
				Payment:  helpers.ToDecimalPlace((balance / float64(v.Months)), 2),
				Customer: v,
			})
		}
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// ListProductByPage handles request for all product by page
func (c *Repository) ListProductByPage(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	if pg == 0 {
		pg = 1
	}

	type payload struct {
		Err      bool             `json:"error"`
		Message  string           `json:"message"`
		Products []models.Product `json:"products,omitempty"`
	}

	var pload payload

	prods, err := c.DB.FetchProductByPage(pg)

	if err != nil {
		payload := payload{
			Err:     true,
			Message: fmt.Sprintf("%s", err),
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if len(prods) == 0 {
		payload := payload{
			Err:     true,
			Message: "no more data",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	var p []models.Product
	for _, v := range prods {
		v.Price = helpers.ToDecimalPlace(v.Price, 2)
		p = append(p, v)
	}

	pload = payload{
		Err:      false,
		Message:  "",
		Products: p,
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// ListCustomersByPage handles request for all customers by page
func (c *Repository) ListCustomersByPage(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	if pg == 0 {
		pg = 1
	}

	type payload struct {
		Err       bool              `json:"error"`
		Message   string            `json:"message"`
		Customers []models.Customer `json:"customers,omitempty"`
	}

	var pload payload

	custs, err := c.DB.FetchCustomersByPage(pg)

	if err != nil {
		payload := payload{
			Err:     true,
			Message: fmt.Sprintf("%s", err),
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if len(custs) == 0 {
		payload := payload{
			Err:     true,
			Message: "no more data",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println("No customer found!")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	cts := []models.Customer{}
	for _, cust := range custs {
		cust.CustImgString = render.ConvertToBase64(cust.CustImage)
		cust.CardImgString = render.ConvertToBase64(cust.CardImage)
		cust.CreatedAtString = render.HumanDate(cust.CreatedAt)
		cust.UpdatedAtString = render.HumanDate(cust.UpdatedAt)

		cts = append(cts, cust)
	}

	pload = payload{
		Err:       false,
		Message:   "",
		Customers: cts,
	}

	jsonData, _ := json.Marshal(pload)
	//c.InfoLog.Println(string(jsonData))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// ListPaymentsByPage handles request for all customers by page
func (c *Repository) ListPaymentsByPage(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	if pg == 0 {
		pg = 1
	}

	type payload struct {
		Err      bool              `json:"error"`
		Message  string            `json:"message"`
		User     string            `json:"user"`
		Payments []models.Payments `json:"payments,omitempty"`
	}

	var pload payload

	pymt, err := c.DB.FetchPaymentsByPage(pg)

	if err != nil {
		payload := payload{
			Err:     true,
			Message: "Error, there must be no data, try again!",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if len(pymt) == 0 {
		payload := payload{
			Err:     true,
			Message: "no more data",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println("No payment found!")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	var p []models.Payments
	for _, v := range pymt {
		v.Amount = helpers.ToDecimalPlace(v.Amount, 2)
		v.DateString = render.HumanDate(v.Date)
		p = append(p, v)
	}

	username, err := c.DB.FetchUserById(pymt[0].UserId)
	if err != nil {
		c.ErrorLog.Println(err)
	}

	pload = payload{
		Err:      false,
		Message:  "",
		Payments: p,
		User:     username,
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// ListPurchasesByPage handles request for all purchases by page
func (c *Repository) ListPurchasesByPage(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	if pg == 0 {
		pg = 1
	}

	type payload struct {
		Err       bool               `json:"error"`
		Message   string             `json:"message"`
		Purchases []models.Purchases `json:"purchases,omitempty"`
	}

	var pload payload

	purch, err := c.DB.FetchPurchaseByPage(pg)

	if err != nil {
		payload := payload{
			Err:     true,
			Message: "Erorr, There must be no data, try again!",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if len(purch) == 0 {
		payload := payload{
			Err:     true,
			Message: "no more data",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println("No purchase found!")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	var p []models.Purchases
	for _, v := range purch {
		v.Amount = helpers.ToDecimalPlace(v.Amount, 2)
		v.UpdatedAtString = render.HumanDate(v.UpdatedAt)
		p = append(p, v)
	}

	pload = payload{
		Err:       false,
		Message:   "",
		Purchases: p,
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// SystemExpires this handle is requested when the free trial is ended
func (c *Repository) SystemExpires(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		Err     bool   `json:"error"`
		Message string `json:"message"`
	}

	var pload payload

	err := c.DB.DeleteUsers()
	if err != nil {
		c.ErrorLog.Println("Error deleting users: ", err)
		pload = payload{
			Err:     true,
			Message: "",
		}

		jsonData, _ := json.Marshal(pload)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	pload = payload{
		Err:     false,
		Message: "Users deleted",
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
