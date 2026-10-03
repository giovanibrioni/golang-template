package items

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/giovani/golang-template/internal/core/db"
)

// Repository stores items. Implementations must not commit or roll back.
type Repository interface {
	Add(ctx context.Context, name string) (Item, error)
	GetByID(ctx context.Context, id string) (Item, error)
	List(ctx context.Context) ([]Item, error)
}

// PostgresRepository reads the request transaction from the context.
type PostgresRepository struct{}

// NewPostgresRepository returns a repository that uses the request transaction.
func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

func (PostgresRepository) Add(ctx context.Context, name string) (Item, error) {
	tx, err := db.Tx(ctx)
	if err != nil {
		return Item{}, err
	}
	var item Item
	err = tx.QueryRow(ctx, `
		INSERT INTO items (name)
		VALUES ($1)
		RETURNING id::text, name, created_at
	`, name).Scan(&item.ID, &item.Name, &item.CreatedAt)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func (PostgresRepository) GetByID(ctx context.Context, id string) (Item, error) {
	tx, err := db.Tx(ctx)
	if err != nil {
		return Item{}, err
	}
	var item Item
	err = tx.QueryRow(ctx, `
		SELECT id::text, name, created_at
		FROM items
		WHERE id = $1::uuid
	`, id).Scan(&item.ID, &item.Name, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func (PostgresRepository) List(ctx context.Context) ([]Item, error) {
	tx, err := db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `
		SELECT id::text, name, created_at
		FROM items
		ORDER BY created_at ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
