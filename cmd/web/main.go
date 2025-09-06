package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jofosuware/small-business-management-app/cmd/web/middleware"
	"github.com/jofosuware/small-business-management-app/cmd/web/routes"
	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/jofosuware/small-business-management-app/internal/driver"
	"github.com/jofosuware/small-business-management-app/internal/handlers"
	"github.com/jofosuware/small-business-management-app/internal/helpers"
	"github.com/jofosuware/small-business-management-app/internal/models"
	"github.com/jofosuware/small-business-management-app/internal/render"
)

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	// Gets port from the platform env
	portNumber := os.Getenv("PORT")

	fmt.Println("Render Port #: ", portNumber)
	if portNumber == "" {
		portNumber = "8080"
	}

	fmt.Printf("Starting frontend server on port %s\n", portNumber)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", portNumber),
		Handler: routes.Routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.User{})
	gob.Register(models.Product{})
	gob.Register(models.Customer{})
	gob.Register(models.Witness{})
	gob.Register([]models.Product{})
	gob.Register(models.Item{})
	gob.Register(models.Payments{})

	//read flags
	inProduction := flag.Bool("production", true, "application is in production")
	//useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "sbma", "Database name")
	dbUser := flag.String("dbuser", "postgres", "Database user")
	dbPass := flag.String("dbpass", "Science@1992", "Database password")
	dbPort := flag.Int("dbport", 5432, "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	//change this to true when in production
	app.InProduction = *inProduction
	app.UseCache = *inProduction

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session
	middleware.Session = session
	middleware.App.InProduction = app.InProduction

	// connect to database
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)

	if *inProduction {
		connectionString = "postgresql://postgres.ilmlvurperawqzzrqbye:09041992Sc!ence@!992$@aws-0-eu-central-1.pooler.supabase.com:5432/sbma"
	}

	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to database!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		errorLog.Println("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = true

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHandlers(&app)

	return db, nil
}
