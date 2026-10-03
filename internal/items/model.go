package items

import "time"

// Item is the public representation of a stored item.
type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// NotFoundError means the item does not exist.
type NotFoundError struct{}

func (NotFoundError) Error() string { return "item not found" }

func (NotFoundError) Kind() string { return "not_found" }

func (NotFoundError) PublicDetail() string { return "Item not found" }

// ErrNotFound is returned when an item id does not match a row.
var ErrNotFound = NotFoundError{}

// ValidationError is a rejected client input. Detail is safe to show.
type ValidationError struct {
	Detail string
}

func (e ValidationError) Error() string { return e.Detail }

func (e ValidationError) Kind() string { return "validation" }

func (e ValidationError) PublicDetail() string { return e.Detail }
