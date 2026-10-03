package items_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	coreerrors "github.com/giovani/golang-template/internal/core/errors"
	"github.com/giovani/golang-template/internal/items"
)

type fakeRepository struct {
	items []items.Item
}

func (f *fakeRepository) Add(_ context.Context, name string) (items.Item, error) {
	item := items.Item{
		ID:        "item-" + name,
		Name:      name,
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	f.items = append(f.items, item)
	return item, nil
}

func (f *fakeRepository) GetByID(_ context.Context, id string) (items.Item, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	return items.Item{}, items.ErrNotFound
}

func (f *fakeRepository) List(context.Context) ([]items.Item, error) {
	if f.items == nil {
		return nil, nil
	}
	out := make([]items.Item, len(f.items))
	copy(out, f.items)
	return out, nil
}

func TestCreate(t *testing.T) {
	repo := &fakeRepository{}
	svc := items.NewService(repo)

	item, err := svc.Create(context.Background(), "widget")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.ID != "item-widget" || item.Name != "widget" {
		t.Fatalf("item = %+v", item)
	}
	if len(repo.items) != 1 {
		t.Fatalf("stored %d items", len(repo.items))
	}
}

func TestList(t *testing.T) {
	repo := &fakeRepository{}
	svc := items.NewService(repo)
	if _, err := svc.Create(context.Background(), "a"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Create(context.Background(), "b"); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 || got[0].Name != "a" || got[1].Name != "b" {
		t.Fatalf("list = %+v", got)
	}
}

func TestListEmpty(t *testing.T) {
	got, err := items.NewService(&fakeRepository{}).List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("list = %#v", got)
	}
}

func TestGet(t *testing.T) {
	svc := items.NewService(&fakeRepository{})
	created, err := svc.Create(context.Background(), "widget")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != created {
		t.Fatalf("got %+v, want %+v", got, created)
	}
}

func TestGetNotFound(t *testing.T) {
	_, err := items.NewService(&fakeRepository{}).Get(context.Background(), "missing")
	if !errors.Is(err, items.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateRejectsName(t *testing.T) {
	tests := []string{"", strings.Repeat("a", 201)}
	for _, name := range tests {
		repo := &fakeRepository{}
		_, err := items.NewService(repo).Create(context.Background(), name)
		var appErr *coreerrors.AppError
		if !errors.As(err, &appErr) || appErr.Code != coreerrors.ErrItemInvalid {
			t.Fatalf("name %q: err = %v", name, err)
		}
		if len(repo.items) != 0 {
			t.Fatalf("name %q stored an item", name)
		}
	}

	repo := &fakeRepository{}
	if _, err := items.NewService(repo).Create(context.Background(), strings.Repeat("a", 200)); err != nil {
		t.Fatalf("200 runes: %v", err)
	}
}
