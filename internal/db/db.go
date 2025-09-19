package db

import (
	"context"
	"database/sql"
	"time"
)

type DBConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

// Creates a new DB connection
func New(dbCfg DBConfig) (*sql.DB, error) {

	// Validate the connection to the address using the postgres driver
	// You must open the connection using a Ping (in this case: at the end of the func)
	db, err := sql.Open("postgres", dbCfg.Addr)
	if err != nil {
		return nil, err
	}

	// Parses the duration: "15m" -> 15 * time.Minute
	parsedMaxIdleTime, err := time.ParseDuration(dbCfg.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	// Sets the attributes for the connection
	db.SetMaxOpenConns(dbCfg.MaxOpenConns)
	db.SetMaxIdleConns(dbCfg.MaxIdleConns)
	db.SetConnMaxIdleTime(parsedMaxIdleTime)

	// Creates a new context with a timeout of 5 seconds to ensure that any database operation
	// using this context does not run indefinitely. The cancel function should be called to release
	// resources once the operation is complete. This is commonly used to prevent resource leaks
	// and to handle cases where the database might be slow or unresponsive.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping the database, creating a new connection and validating it
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
