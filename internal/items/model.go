package items

import (
	"time"

	"github.com/giovanibrioni/golang-template/internal/core/errors"
)

// Item is the public representation of a stored item.
type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrNotFound is returned when an item id does not match a row.
var ErrNotFound = errors.NewItemNotFound("")

// ErrInvalid is returned when item data is invalid.
var ErrInvalid = errors.NewItemInvalid("")

// ErrConflict is returned when an item already exists.
var ErrConflict = errors.NewItemConflict("")
