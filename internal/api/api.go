package api

import (
	"log"
	"net/http"
	"time"

	"github.com/dottox/social/internal/db"
	"github.com/dottox/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	Config Config
	Store  store.Storage
}

type Config struct {
	Addr    string
	Env     string
	Version string
	DB      db.DBConfig
}

// Mount functions allow the app to create their router
// In this case we are using chi. But you can use any other Handler
// Like the standard mux one.
// Chi is good because allow route grouping and ease of middleware use
func (app *Application) Mount() http.Handler {

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
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Get("/", app.getPostHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)

				r.Route("/comments", func(r chi.Router) {
					r.Post("/", app.createCommentHandler)
					r.Get("/", app.getCommentsByPostHandler)
				})
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Post("/", app.createUserHandler)
		})
	})

	return r
}

// Run function allow to create and run the server
func (app *Application) Run(mux http.Handler) error {

	// creates the server with the application config
	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("starting server on %s", app.Config.Addr)

	// start the server
	return srv.ListenAndServe()
}
