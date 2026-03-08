package db

import (
	"database/sql"
	"utils/types"

	_ "github.com/lib/pq"
)

// Connect establishes a connection to the database.
func Connect(uri string) types.Result[*sql.DB, error] {
	return types.Pipe3(
		types.Ok[string, error](uri),
		func(uri string) types.Result[*sql.DB, error] {
			db, err := sql.Open("postgres", uri)
			if err != nil {
				return types.Err[*sql.DB, error](err)
			}
			return types.Ok[*sql.DB, error](db)
		},
		func(db *sql.DB) types.Result[*sql.DB, error] {
			if err := db.Ping(); err != nil {
				return types.Err[*sql.DB, error](err)
			}
			return types.Ok[*sql.DB, error](db)
		},
		func(db *sql.DB) *sql.DB {
			return db
		},
	)
}
