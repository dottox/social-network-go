package main

import (
	"log"

	"github.com/dottox/social/internal/api"
	"github.com/dottox/social/internal/db"
	"github.com/dottox/social/internal/env"
	"github.com/dottox/social/internal/store"
)

func main() {
	// Loads enviromental variables
	env.LoadEnvs()

	// Creates the DBConfiguration for the connection pool
	dbCfg := db.DBConfig{
		Addr:         env.GetString("DB_ADDR", "postgres://postgres:admin@localhost:5432/social_go?sslmode=disable"),
		MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
		MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	// Creates the api configuration, containing the DBConfig
	cfg := api.Config{
		Addr:    env.GetString("ADDR", ":8080"),
		Env:     env.GetString("ENV", "development"),
		Version: env.GetString("VERSION", "x.x.x"),
		DB:      dbCfg,
	}

	// Create a new DB connection with the DBConfig
	db, err := db.New(dbCfg)
	if err != nil {
		log.Panic(err)
	}

	// Defer the Close function to close the database at the end
	defer db.Close()

	log.Println("database connection pool established")

	// Create a new Storage
	// Storage is a struct containing all the repositories (stores)
	store := store.NewStorage(db)

	// Create a new application
	app := &api.Application{
		Config: cfg,
		Store:  store,
	}

	// Create a new router, in this case we are using Chi
	// We're gonna run the app with the router
	router := app.Mount()

	// Run the app and Fatal if any errors.
	log.Fatal(app.Run(router))
}
