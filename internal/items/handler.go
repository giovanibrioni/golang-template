package items

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/giovani/golang-template/internal/core/db"
	"github.com/giovani/golang-template/internal/core/errors"
	"github.com/giovani/golang-template/internal/core/httpx"
)

// Mount registers item routes on r. r is already prefixed with /api/v1/items.
func Mount(r chi.Router, beginner db.Beginner, svc *Service) {
	r.With(httpx.Validate[CreateItemRequest]).Post("/", httpx.Wrap(beginner, create(svc)))
	r.Get("/", httpx.Wrap(beginner, list(svc)))
	r.Get("/{id}", httpx.Wrap(beginner, get(svc)))
}

// CreateItemRequest is the request body for creating an item.
type CreateItemRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

// CreateItemResponse is the response for creating an item.
type CreateItemResponse struct {
	Item Item `json:"item"`
}

// ListItemsResponse is the response for listing items.
type ListItemsResponse struct {
	Items []Item `json:"items"`
}

// GetItemResponse is the response for getting an item.
type GetItemResponse struct {
	Item Item `json:"item"`
}

// create handles POST /api/v1/items.
// @Summary Create a new item
// @Description Create a new item with the given name
// @Tags items
// @Accept json
// @Produce json
// @Param request body CreateItemRequest true "Item to create"
// @Success 201 {object} CreateItemResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/v1/items [post]
func create(svc *Service) httpx.Handler {
	return func(r *http.Request) (httpx.Response, error) {
		req, ok := httpx.ValidatedBody[CreateItemRequest](r)
		if !ok {
			return httpx.Response{}, errors.New(errors.ErrValidation, errors.WithDetail("Validation skipped"))
		}
		item, err := svc.Create(r.Context(), req.Name)
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Status: http.StatusCreated, Body: CreateItemResponse{Item: item}}, nil
	}
}

// list handles GET /api/v1/items.
// @Summary List all items
// @Description Get a list of all items
// @Tags items
// @Produce json
// @Success 200 {object} ListItemsResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/v1/items [get]
func list(svc *Service) httpx.Handler {
	return func(r *http.Request) (httpx.Response, error) {
		items, err := svc.List(r.Context())
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Status: http.StatusOK, Body: ListItemsResponse{Items: items}}, nil
	}
}

// get handles GET /api/v1/items/{id}.
// @Summary Get an item by ID
// @Description Get a single item by its ID
// @Tags items
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} GetItemResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/v1/items/{id} [get]
func get(svc *Service) httpx.Handler {
	return func(r *http.Request) (httpx.Response, error) {
		item, err := svc.Get(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Status: http.StatusOK, Body: GetItemResponse{Item: item}}, nil
	}
}
