package main

import (
	"expvar"
	"runtime"
	"time"

	"github.com/dottox/social/internal/api"
	"github.com/dottox/social/internal/auth"
	"github.com/dottox/social/internal/db"
	"github.com/dottox/social/internal/env"
	"github.com/dottox/social/internal/mailer"
	"github.com/dottox/social/internal/ratelimiter"
	"github.com/dottox/social/internal/store"
	"github.com/dottox/social/web"
	"go.uber.org/zap"
)

//	@title			GopherSocial API
//	@description	API for Gopher Social

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

// @BasePath					/v1
// @schemes					http
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	// Loads enviromental variables
	env.LoadEnvs()

	// Creates a new zap logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Creates the DBConfiguration for the connection pool
	dbCfg := db.DBConfig{
		Addr:         env.GetString("DB_ADDR", ""),
		MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
		MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	// Creates the api configuration, containing the DBConfig
	cfg := api.Config{
		Protocol:    env.GetString("PROTOCOL", "http"),
		Addr:        env.GetString("ADDR", "localhost"),
		Port:        env.GetString("PORT", ":8080"),
		FrontendURL: env.GetString("FRONTEND_URL", "http://localhost:8080"),
		Env:         env.GetString("ENV", "development"),
		Version:     env.GetString("VERSION", "x.x.x"),
		DB:          dbCfg,
		Mail: api.MailConfig{
			Exp:       time.Hour * 24, // 24 hours
			FromEmail: env.GetString("FROM_EMAIL", ""),
			SendGrid: api.SendGridConfig{
				APIKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		Auth: api.AuthConfig{
			Basic: api.BasicConfig{
				Username: env.GetString("BASIC_AUTH_USERNAME", "admin"),
				Password: env.GetString("BASIC_AUTH_PASSWORD", "admin"),
			},
			Token: api.TokenConfig{
				Secret: env.GetString("AUTH_TOKEN_SECRET", "secret"),
				Exp:    time.Hour * 24 * 3, // 24 hours
				Iss:    "gophersocial",
			},
		},
		RateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: 50,
			TimeFrame:            time.Minute,
			Enabled:              true,
		},
	}

	// Create a new DB connection with the DBConfig
	db, err := db.New(dbCfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer the Close function to close the database at the end
	defer db.Close()

	logger.Info("database connection pool established")

	// Create a new Storage
	// Storage is a struct containing all the repositories (stores)
	store := store.NewStorage(db)

	mailer := mailer.NewSendGridMailer(cfg.Mail.SendGrid.APIKey, cfg.Mail.FromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.Auth.Token.Secret,
		cfg.Auth.Token.Iss,
		cfg.Auth.Token.Iss,
	)

	rateLimiter := ratelimiter.NewFixedWindowRateLimiter(
		cfg.RateLimiter.RequestsPerTimeFrame,
		cfg.RateLimiter.TimeFrame,
	)

	// Create a new application
	app := &api.Application{
		Config:        cfg,
		Store:         store,
		Logger:        logger,
		Mailer:        mailer,
		Authenticator: jwtAuthenticator,
		RateLimiter:   rateLimiter,
	}

	// Publish some metrics to /v1/metrics
	expvar.NewString("version").Set(cfg.Version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// Create a new router, in this case we are using Chi
	// We're gonna run the app with the router
	router := app.Mount()

	// Start the web app for serving static files and the index page
	apiUrl := env.GetString("API_URL", "http://localhost:8080")
	webApp := web.NewWebApp(
		env.GetString("FRONTEND_PROTOCOL", "http"),
		env.GetString("FRONTEND_ADDR", "localhost"),
		env.GetString("FRONTEND_PORT", ":4000"),
		apiUrl,
	)
	webRouter := webApp.Mount()
	go func() {
		err := webApp.Run(webRouter)
		if err != nil {
			logger.Fatal(err)
		}
	}()

	// Run the app and Fatal if any errors.
	logger.Fatal(app.Run(router))
}
