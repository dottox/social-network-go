package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dottox/social/docs"
	"github.com/dottox/social/internal/auth"
	"github.com/dottox/social/internal/db"
	"github.com/dottox/social/internal/mailer"
	"github.com/dottox/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type Application struct {
	Config        Config
	Store         store.Storage
	Logger        *zap.SugaredLogger
	Mailer        mailer.Client
	Authenticator auth.Authenticator
}

type Config struct {
	Protocol    string
	Addr        string
	Port        string
	FrontendURL string
	Env         string
	Version     string
	Mail        MailConfig
	DB          db.DBConfig
	Auth        AuthConfig
}

type AuthConfig struct {
	Basic BasicConfig
	Token TokenConfig
}

type BasicConfig struct {
	Username string
	Password string
}

type TokenConfig struct {
	Secret string
	Exp    time.Duration
	Iss    string
}

type MailConfig struct {
	SendGrid  SendGridConfig
	Exp       time.Duration
	FromEmail string
}

type SendGridConfig struct {
	APIKey string
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
	docsURL := fmt.Sprintf("%s://%s%s/swagger/doc.json", app.Config.Protocol, app.Config.Addr, app.Config.Port)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

	// Define routes, you can have subroutes
	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)

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
			r.Use(app.AuthTokenMiddleware)

			r.Route("/{userId}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.getTokenHandler)
			r.Put("/user/activate", app.activateUserHandler)
		})
	})

	return r
}

// Run function allow to create and run the server
func (app *Application) Run(mux http.Handler) error {

	// docs
	docs.SwaggerInfo.Version = app.Config.Version
	docs.SwaggerInfo.Host = app.Config.Addr + app.Config.Port
	docs.SwaggerInfo.BasePath = "/v1"

	// creates the server with the application config
	srv := &http.Server{
		Addr:         app.Config.Addr + app.Config.Port,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	app.Logger.Infow("starting server", "protocol", app.Config.Protocol, "addr", srv.Addr, "env", app.Config.Env)

	// start the server
	return srv.ListenAndServe()
}
