package application

import (
	"log"

	"github.com/jofosuware/small-business-management-app/internal/repository"
)

type App struct {
	Port     int
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	DB       repository.DatabaseRepo
}
