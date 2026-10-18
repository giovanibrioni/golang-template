package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/giovanibrioni/golang-template/internal/core/db"
	"github.com/giovanibrioni/golang-template/internal/core/httpx"
	"github.com/giovanibrioni/golang-template/internal/health"
	"github.com/giovanibrioni/golang-template/internal/items"
)

// Mount registers all v1 routes on r.
func Mount(r chi.Router, beginner db.Beginner, itemsService *items.Service) {
	r.Get("/health", httpx.Wrap(beginner, health.Handler))
	r.Route("/items", func(r chi.Router) {
		items.Mount(r, beginner, itemsService)
	})
}

// Routes returns an http.Handler with all v1 routes mounted under /api/v1.
func Routes(beginner db.Beginner, itemsService *items.Service) http.Handler {
	r := chi.NewRouter()
	Mount(r, beginner, itemsService)
	return r
}
