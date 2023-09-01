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
		Err     bool   `json:"error"`
		Message string `json:"message"`
		Debt    int    `json:"debt,omitempty"`
	}

	balance, err := c.CalcCustomerDebt(custId)
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

	pload := payload{
		Err:     false,
		Message: "",
		Debt:    balance,
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// CalcCustomerDebt calculates customer debt to the enterprise
func (c *Repository) CalcCustomerDebt(customerId string) (int, error) {
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

	amount := 0
	if len(custPymt) != 0 {
		for _, v := range custPymt {
			amount += int(v.Amount)
		}
	}

	balance := 0
	for _, v := range custDebt {
		balance += int(v.Balance)
	}

	balance -= amount
	return balance, nil
}

// CustomerOwingToday handles the request for the customers owing at the present day
func (c *Repository) CustomerOwingToday(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		Err      bool            `json:"error"`
		Message  string          `json:"message"`
		Debt     int             `json:"debt,omitempty"`
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

		if presentDay == pastDay {
			balance, _ := c.CalcCustomerDebt(v.CustomerId)

			pload = append(pload, payload{
				Err:      false,
				Message:  "",
				Debt:     balance,
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
			Err:     false,
			Message: "no more data",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	pload = payload{
		Err:      false,
		Message:  "",
		Products: prods,
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
			Err:     false,
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
		Payments []models.Payments `json:"payments,omitempty"`
	}

	var pload payload

	pymt, err := c.DB.FetchPaymentsByPage(pg)

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

	if len(pymt) == 0 {
		payload := payload{
			Err:     false,
			Message: "no more data",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println("No payment found!")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	pload = payload{
		Err:      false,
		Message:  "",
		Payments: pymt,
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
			Message: fmt.Sprintf("%s", err),
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if len(purch) == 0 {
		payload := payload{
			Err:     false,
			Message: "no more data",
		}

		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println("No purchase found!")
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	pload = payload{
		Err:       false,
		Message:   "",
		Purchases: purch,
	}

	jsonData, _ := json.Marshal(pload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
