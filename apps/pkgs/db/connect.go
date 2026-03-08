package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// Connect establishes a connection to the database.
func Connect(uri string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	return conn, nil
}
