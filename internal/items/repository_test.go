package items_test

import (
	"errors"
	"os"
	"sort"
	"testing"

	"github.com/giovanibrioni/golang-template/internal/core/pgtest"
	"github.com/giovanibrioni/golang-template/internal/items"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Run(m))
}

func TestPostgresRepository(t *testing.T) {
	pool := pgtest.Require(t)
	ctx := pgtest.TxContext(t, pool)
	repo := items.NewPostgresRepository()

	first, err := repo.Add(ctx, "first")
	if err != nil {
		t.Fatalf("add first: %v", err)
	}
	second, err := repo.Add(ctx, "second")
	if err != nil {
		t.Fatalf("add second: %v", err)
	}
	if first.ID == "" || second.ID == "" || first.ID == second.ID {
		t.Fatalf("ids = %q %q", first.ID, second.ID)
	}
	if first.CreatedAt.IsZero() || second.CreatedAt.IsZero() {
		t.Fatal("created_at is zero")
	}

	got, err := repo.GetByID(ctx, first.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != first.ID || got.Name != "first" {
		t.Fatalf("got %+v", got)
	}

	listed, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := []string{first.ID, second.ID}
	sort.Strings(want)
	if len(listed) != 2 || listed[0].ID != want[0] || listed[1].ID != want[1] {
		t.Fatalf("list = %+v, want ids %v", listed, want)
	}

	_, err = repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, items.ErrNotFound) {
		t.Fatalf("missing id: %v", err)
	}
	_, err = repo.GetByID(ctx, "not-a-uuid")
	if !errors.Is(err, items.ErrNotFound) {
		t.Fatalf("invalid id: %v", err)
	}
}
