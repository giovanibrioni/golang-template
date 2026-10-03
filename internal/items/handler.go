package items

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/giovani/golang-template/internal/core/db"
	"github.com/giovani/golang-template/internal/core/httpx"
)

// Mount registers item routes on r. r is already prefixed with /api/v1/items.
func Mount(r chi.Router, beginner db.Beginner, svc *Service) {
	r.Post("/", httpx.Wrap(beginner, create(svc)))
	r.Get("/", httpx.Wrap(beginner, list(svc)))
	r.Get("/{id}", httpx.Wrap(beginner, get(svc)))
}

func create(svc *Service) httpx.Handler {
	return func(r *http.Request) (httpx.Response, error) {
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return httpx.Response{}, ValidationError{Detail: "Invalid JSON"}
		}
		item, err := svc.Create(r.Context(), body.Name)
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Status: http.StatusCreated, Body: item}, nil
	}
}

func list(svc *Service) httpx.Handler {
	return func(r *http.Request) (httpx.Response, error) {
		items, err := svc.List(r.Context())
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Status: http.StatusOK, Body: items}, nil
	}
}

func get(svc *Service) httpx.Handler {
	return func(r *http.Request) (httpx.Response, error) {
		item, err := svc.Get(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Status: http.StatusOK, Body: item}, nil
	}
}
