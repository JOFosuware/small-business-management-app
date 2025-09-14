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
	//read flags
	inProduction := flag.Bool("production", true, "application is in production")
	//useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.Int("dbport", 5432, "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	// Gets port from the platform env
	portNumber := os.Getenv("PORT")
	fmt.Println("Render Port #: ", portNumber)
	if portNumber == "" {
		portNumber = "8081"
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// connect to database
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)

	if *inProduction {
		connectionString = "postgres://postgres.ilmlvurperawqzzrqbye:$@aws-0-eu-central-1.pooler.supabase.com/sbma"
	}

	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to database!")

	d := db.SQL

	//Initialize api handler
	model := dbrepo.NewDB(d)
	apihandler.Repo = apihandler.Repository{
		DB:       model,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", portNumber),
		Handler:           apiRoutes.Routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	infoLog.Printf("Starting Back end server on port %s\n", portNumber)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
