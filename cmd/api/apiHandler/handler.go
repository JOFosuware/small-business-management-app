package apihandler

import (
	"encoding/json"
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

	if custId == "" {
		payload := payload{
			Err:     true,
			Message: "customerId is not provided in the url",
		}
		jsonData, _ := json.Marshal(payload)
		c.InfoLog.Println(string(jsonData))
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	custDebt, err := c.DB.CustomerDebt(custId)
	if err != nil {
		payload := payload{
			Err:     true,
			Message: "customer debt information can't be retrieved",
		}
		jsonData, _ := json.Marshal(payload)
		c.ErrorLog.Println(err)
		c.InfoLog.Println(string(jsonData))
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	if len(custDebt) == 0 {
		payload := payload{
			Err:     true,
			Message: fmt.Sprintf("Customer with this id: %s does not owe!", custId),
		}
		jsonData, _ := json.Marshal(payload)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	balance := 0
	for _, v := range custDebt {
		balance += int(v.Balance)
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
