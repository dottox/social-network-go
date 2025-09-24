package main

import (
	"log"

	"github.com/dottox/social/internal/db"
	"github.com/dottox/social/internal/env"
	"github.com/dottox/social/internal/store"
)

func main() {
	err := env.LoadEnvs()
	if err != nil {
		log.Fatal(err)
	}
	dbAddr := env.GetString("DB_ADDR", "")
	conn, err := db.New(db.DBConfig{dbAddr, 15, 15, "15m"})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	myStore := store.NewStorage(conn)

	err = db.Seed(myStore, conn)
	if err != nil {
		log.Fatal(err)
	}
}
