// Package api wires the HTTP router.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/giovani/golang-template/internal/core/db"
	"github.com/giovani/golang-template/internal/core/httpx"
	"github.com/giovani/golang-template/internal/health"
	"github.com/giovani/golang-template/internal/items"
)

// NewRouter mounts versioned routes and the shared middleware.
func NewRouter(beginner db.Beginner, itemsService *items.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(httpx.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", httpx.Wrap(beginner, health.Handler))
		r.Route("/items", func(r chi.Router) {
			items.Mount(r, beginner, itemsService)
		})
	})
	return r
}
