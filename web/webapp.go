package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/dottox/social/web/components"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type (
	WebApp struct {
		config webConfig
		apiUrl string
	}

	webConfig struct {
		protocol string
		domain   string
		port     string
	}
)

func NewWebApp(protocol, domain, port, apiUrl string) *WebApp {
	return &WebApp{
		config: webConfig{
			protocol: protocol,
			domain:   domain,
			port:     port,
		},
		apiUrl: apiUrl,
	}
}

func (app *WebApp) Mount() http.Handler {

	// Creates a new chi router
	r := chi.NewRouter()

	// Assign the router to use these middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Timeout for the middlewares
	r.Use(middleware.Timeout(60 * time.Second))

	// Define routes, you can have subroutes
	r.Get("/", templ.Handler(components.Index()).ServeHTTP)
	r.Get("/feed", app.FeedHandler)
	r.Get("/activate", app.ActivateUserHandler)
	return r
}

func (app *WebApp) Run(mux http.Handler) error {
	// creates the server with the application config
	srv := &http.Server{
		Addr:         app.config.domain + app.config.port,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	fmt.Printf("Starting web server on %s://%s%s\n", app.config.protocol, app.config.domain, app.config.port)

	// start the server
	return srv.ListenAndServe()

}
