package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"utils/db/db"

	_ "github.com/lib/pq"
)

// SetupTestTx creates a transaction-wrapped Querier for testing.
// The transaction is automatically rolled back when the test completes,
// ensuring test isolation without leaving data in the database.
func SetupTestTx(t *testing.T) db.Querier {
	t.Helper()

	dsn := os.Getenv("DB_URI")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@db:5432/test?sslmode=disable"
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		t.Fatalf("failed to ping database: %v", err)
	}

	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		conn.Close()
		t.Fatalf("failed to begin transaction: %v", err)
	}

	t.Cleanup(func() {
		tx.Rollback()
		conn.Close()
	})

	return db.New(tx)
}
