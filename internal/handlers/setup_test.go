package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"

	"github.com/go-course/bookings/internal/config"
	"github.com/go-course/bookings/internal/helpers"
	"github.com/go-course/bookings/internal/models"
	"github.com/go-course/bookings/internal/render"
)

var (
	app                 config.AppConfig
	session             *scs.SessionManager
	pathToTemplateCache = "./../../templates"
	infoLog             *log.Logger
	errorLog            *log.Logger
)

var functions = template.FuncMap{
	"humanDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	"formatDate": func(t time.Time, format string) string {
		return t.Format(format)
	},
	// Iterate returns a slice of ints, starting at 1 and going to count
	"iterate": func(count int) []int {
		var items []int
		for i := 0; i < count; i++ {
			items = append(items, i)
		}
		return items
	},
	"add": func(a, b int) int {
		return a + b
	},
}

func TestMain(m *testing.M) {
	// what am I going to put in the session
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.Reservation{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	// change this to true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)
	listenForMail()

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal(err.Error())
	}

	app.TemplateCache = tc
	app.UseCache = true

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	render.NewRender(&app)
	helpers.NewHelpers(&app)
	os.Exit(m.Run())
}

func listenForMail() {
	go func() {
		for {
			<-app.MailChan
		}
	}()
}

func getRoutes() http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/contact", Repo.Contact)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)

	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", Repo.ChooseRoom)
	mux.Get("/book-room", Repo.BookRoom)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		// mux.Use(Auth)
		mux.Get("/dashboard", Repo.AdminDashboard)
		mux.Get("/reservations-new", Repo.AdminNewReservations)
		mux.Get("/reservations-all", Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", Repo.AdminPostReservationsCalendar)

		mux.Get("/reservations/{src}/{id}/show", Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", Repo.AdminPostShowReservation)
		mux.Get("/process-reservation/{src}/{id}/do", Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", Repo.AdminDeleteReservation)

	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}

// NoSurf is the csrf protection middleware
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves session data for current request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache creates a template cache as a map
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := make(map[string]*template.Template)
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplateCache))
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplateCache))
		if err != nil {
			return nil, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplateCache))
			if err != nil {
				return nil, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}
