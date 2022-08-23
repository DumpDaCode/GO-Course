package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-course/bookings/internal/config"
	"github.com/go-course/bookings/internal/handlers"
	"github.com/go-course/bookings/internal/helpers"
	"github.com/go-course/bookings/internal/models"
	"github.com/go-course/bookings/internal/render"
)

const portNumber = ":8080"

var (
	app      config.AppConfig
	session  *scs.SessionManager
	infoLog  *log.Logger
	errorLog *log.Logger
)

// main is the main application function
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Starting application on Port %s\n", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error {
	//what am I going to store
	gob.Register(models.Reservation{})

	// Change this to true when in production
	app.InProduction = false

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

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	return nil
}
