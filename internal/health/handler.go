// Package health serves the liveness probe.
package health

import (
	"net/http"

	"github.com/giovani/golang-template/internal/core/db"
	"github.com/giovani/golang-template/internal/core/httpx"
)

type statusResponse struct {
	Status string `json:"status"`
}

// Handler checks that the request transaction can run a query.
func Handler(r *http.Request) (httpx.Response, error) {
	tx, err := db.Tx(r.Context())
	if err != nil {
		return httpx.Response{}, err
	}
	var one int
	if err := tx.QueryRow(r.Context(), "SELECT 1").Scan(&one); err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{
		Status: http.StatusOK,
		Body:   statusResponse{Status: "ok"},
	}, nil
}
