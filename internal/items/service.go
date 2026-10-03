package items

import (
	"context"
	"log/slog"
	"unicode/utf8"
)

// Service applies item rules. It does not know about HTTP or transactions.
type Service struct {
	repo Repository
}

// NewService returns a service that uses repo.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create stores a new item. Name must contain between 1 and 200 runes.
func (s *Service) Create(ctx context.Context, name string) (Item, error) {
	if name == "" || utf8.RuneCountInString(name) > 200 {
		return Item{}, ValidationError{Detail: "Name must be between 1 and 200 characters"}
	}
	item, err := s.repo.Add(ctx, name)
	if err != nil {
		return Item{}, err
	}
	slog.InfoContext(ctx, "item_created", "item_id", item.ID)
	return item, nil
}

// Get returns the item or ErrNotFound when the repository reports it is missing.
func (s *Service) Get(ctx context.Context, id string) (Item, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

// List returns every item. An empty result is a non-nil slice.
func (s *Service) List(ctx context.Context) ([]Item, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []Item{}, nil
	}
	return items, nil
}
