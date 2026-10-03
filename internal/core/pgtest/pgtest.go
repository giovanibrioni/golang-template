// Package pgtest connects integration tests to Postgres and rolls back their data.
package pgtest

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/giovani/golang-template/internal/core/db"
)

// Pool is set by Run when DATABASE_URL is available.
var Pool *pgxpool.Pool

// Run migrates the test database once, runs tests, and closes the pool.
// Short mode and a missing DATABASE_URL skip the database entirely.
func Run(m *testing.M) int {
	flag.Parse()
	if testing.Short() || os.Getenv("DATABASE_URL") == "" {
		return m.Run()
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		return 1
	}
	if err := db.Migrate(ctx, pool); err != nil {
		pool.Close()
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		return 1
	}

	Pool = pool
	code := m.Run()
	pool.Close()
	Pool = nil
	return code
}

// Require skips the test when Postgres is not available.
func Require(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if Pool == nil {
		t.Skip("DATABASE_URL is not set")
	}
	return Pool
}

// RollbackBeginner starts an outer transaction and returns a beginner whose
// Begin opens a savepoint. Cleanup rolls the outer transaction back, so a
// commit inside httpx.Wrap does not persist.
func RollbackBeginner(t *testing.T, pool *pgxpool.Pool) db.Beginner {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})
	return savepointBeginner{tx: tx}
}

// TxContext starts a transaction and stores it on the context for repository tests.
// Cleanup rolls it back.
func TxContext(t *testing.T, pool *pgxpool.Pool) context.Context {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})
	return db.WithTx(ctx, tx)
}

type savepointBeginner struct {
	tx pgx.Tx
}

func (s savepointBeginner) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.tx.Begin(ctx)
}
