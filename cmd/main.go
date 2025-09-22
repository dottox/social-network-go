package main

import (
	"github.com/dottox/social/internal/api"
	"github.com/dottox/social/internal/db"
	"github.com/dottox/social/internal/env"
	"github.com/dottox/social/internal/store"
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
		Protocol: env.GetString("PROTOCOL", "http"),
		Addr:     env.GetString("ADDR", "localhost"),
		Port:     env.GetString("PORT", ":8080"),
		Env:      env.GetString("ENV", "development"),
		Version:  env.GetString("VERSION", "x.x.x"),
		DB:       dbCfg,
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

	// Create a new application
	app := &api.Application{
		Config: cfg,
		Store:  store,
		Logger: logger,
	}

	// Create a new router, in this case we are using Chi
	// We're gonna run the app with the router
	router := app.Mount()

	// Run the app and Fatal if any errors.
	logger.Fatal(app.Run(router))
}
