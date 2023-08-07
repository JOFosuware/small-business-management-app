package apihandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
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
