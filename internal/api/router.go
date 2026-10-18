// Package api wires the HTTP router.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/giovanibrioni/golang-template/internal/api/v1"
	"github.com/giovanibrioni/golang-template/internal/core/db"
	"github.com/giovanibrioni/golang-template/internal/core/httpx"
	"github.com/giovanibrioni/golang-template/internal/health"
	"github.com/giovanibrioni/golang-template/internal/items"
)

// NewRouter mounts versioned routes and the shared middleware.
// Versioning strategy: URL-based (/api/v1, /api/v2).
// Add new versions by creating internal/api/vN and mounting here.
func NewRouter(beginner db.Beginner, itemsService *items.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(httpx.RequestID)
	r.Use(httpx.OTelMiddleware)

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health", httpx.Wrap(beginner, health.Handler))

	// Mount v1 routes
	r.Mount("/api/v1", v1.Routes(beginner, itemsService))

	// Future: mount v2 routes
	// r.Mount("/api/v2", v2.Routes(beginner, itemsService))

	return r
}
