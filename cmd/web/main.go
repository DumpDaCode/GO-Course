package main

import (
	"fmt"

	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-course/bookings/internal/config"
	"github.com/go-course/bookings/internal/handlers"
	"github.com/go-course/bookings/internal/render"
)

const portNumber = ":8080"

var (
	app     config.AppConfig
	session *scs.SessionManager
)

// main is the main application function
func main() {

	// Change this to true when in production
	app.InProduction = false

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

	render.NewTemplates(&app)

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	fmt.Printf("Starting application on Port %s\n", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}
