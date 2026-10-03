package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ctxKey string

const validatedKey ctxKey = "validated"

var validate = validator.New(validator.WithRequiredStructEnabled())

func ValidatedBody[T any](r *http.Request) (T, bool) {
	v, ok := r.Context().Value(validatedKey).(T)
	return v, ok
}

func Validate[T any](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req T
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.InfoContext(r.Context(), "request_invalid", "error", "invalid JSON", "status", http.StatusBadRequest)
			writeJSON(w, http.StatusBadRequest, errorBody{Detail: "Invalid JSON"})
			return
		}
		if err := validate.Struct(req); err != nil {
			var ve validator.ValidationErrors
			if errorsAs(err, &ve) {
				detail := formatValidationErrors(ve)
				slog.InfoContext(r.Context(), "request_invalid", "error", detail, "status", http.StatusBadRequest)
				writeJSON(w, http.StatusBadRequest, errorBody{Detail: detail})
				return
			}
			slog.InfoContext(r.Context(), "request_invalid", "error", "validation failed", "status", http.StatusBadRequest)
			writeJSON(w, http.StatusBadRequest, errorBody{Detail: "Validation failed"})
			return
		}
		ctx := context.WithValue(r.Context(), validatedKey, req)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func formatValidationErrors(ve validator.ValidationErrors) string {
	if len(ve) == 0 {
		return "Validation failed"
	}
	return ve[0].Error()
}

func errorsAs(err error, target any) bool {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return true
	}
	return false
}
