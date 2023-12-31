package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	apihandler "github.com/jofosuware/small-business-management-app/cmd/api/apiHandler"
	"github.com/jofosuware/small-business-management-app/cmd/api/apiRoutes"
	"github.com/jofosuware/small-business-management-app/internal/driver"
	"github.com/jofosuware/small-business-management-app/internal/repository/dbrepo"
)

func main() {
	// read flags
	serverPort := flag.Int("port", 8081, "Port the server is starting on")
	dbHost := flag.String("dbhost", "dpg-cm8k2di1hbls73acuh4g-a", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// connect to database
	log.Println("\nConnecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("\nConnected to database!")

	d := db.SQL

	//Initialize api handler
	model := dbrepo.NewDB(d)
	apihandler.Repo = apihandler.Repository{
		DB:       model,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", *serverPort),
		Handler:           apiRoutes.Routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	infoLog.Printf("\nStarting Back end server on port %d\n", *serverPort)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
