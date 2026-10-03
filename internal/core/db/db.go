// Package db owns the Postgres pool, request transactions, and migrations.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// Beginner starts a transaction. *pgxpool.Pool implements it.
// Tests pass a beginner whose Begin opens a savepoint on an outer transaction.
type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// NewPool opens a pool and checks that Postgres accepts connections.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// WithTx stores the request transaction on the context.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// Tx returns the transaction opened for the current request.
func Tx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok || tx == nil {
		return nil, errors.New("transaction missing from context")
	}
	return tx, nil
}
