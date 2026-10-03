// Package httpx adapts net/http handlers that return a value and an error.
// It owns the request transaction and the mapping from domain errors to HTTP status.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/giovani/golang-template/internal/core/db"
	"github.com/giovani/golang-template/internal/core/logging"
)

// Response is the HTTP result of a handler. The wrapper writes it after commit.
type Response struct {
	Status int
	Body   any
}

// Handler handles one request. It must not write to the ResponseWriter.
type Handler func(*http.Request) (Response, error)

type publicError interface {
	error
	Kind() string
	PublicDetail() string
}

type errorBody struct {
	Detail string `json:"detail"`
}

// RequestID assigns a random id, stores it on the request context, and echoes X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newRequestID()
		w.Header().Set("X-Request-ID", id)
		ctx := logging.WithRequestID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Wrap opens one transaction for the request.
// The handler's repository must use that transaction. Commit and rollback stay here.
func Wrap(beginner db.Beginner, handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tx, err := beginner.Begin(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		ctx := db.WithTx(r.Context(), tx)
		resp, err := handler(r.WithContext(ctx))
		if err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			writeError(w, r, err)
			return
		}
		if err := tx.Commit(ctx); err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, resp.Status, resp.Body)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, detail := mapError(r.Context(), err)
	writeJSON(w, status, errorBody{Detail: detail})
}

func mapError(ctx context.Context, err error) (int, string) {
	var public publicError
	if errors.As(err, &public) {
		switch public.Kind() {
		case "not_found":
			return http.StatusNotFound, "Item not found"
		case "validation":
			detail := public.PublicDetail()
			if detail == "" {
				detail = "Invalid request"
			}
			return http.StatusBadRequest, detail
		}
	}
	slog.ErrorContext(ctx, "request_failed", "error", err)
	return http.StatusInternalServerError, "Internal server error"
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write_json", "error", err)
	}
}

func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buf[:])
}
