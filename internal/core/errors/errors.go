package errors

import "net/http"

// ErrorCode is a machine-readable error identifier.
type ErrorCode string

const (
	// Generic
	ErrInternal     ErrorCode = "internal_error"
	ErrNotFound     ErrorCode = "not_found"
	ErrValidation   ErrorCode = "validation_error"
	ErrUnauthorized ErrorCode = "unauthorized"
	ErrForbidden    ErrorCode = "forbidden"
	ErrConflict     ErrorCode = "conflict"
	ErrRateLimited  ErrorCode = "rate_limited"
	ErrUnavailable  ErrorCode = "service_unavailable"

	// Items domain
	ErrItemNotFound ErrorCode = "item_not_found"
	ErrItemInvalid  ErrorCode = "item_invalid"
	ErrItemConflict ErrorCode = "item_conflict"
)

// Catalog maps error codes to HTTP status and public message.
type CatalogEntry struct {
	Code    ErrorCode
	Status  int
	Message string
}

var catalog = map[ErrorCode]CatalogEntry{
	ErrInternal:     {ErrInternal, http.StatusInternalServerError, "Internal server error"},
	ErrNotFound:     {ErrNotFound, http.StatusNotFound, "Resource not found"},
	ErrValidation:   {ErrValidation, http.StatusBadRequest, "Invalid request"},
	ErrUnauthorized: {ErrUnauthorized, http.StatusUnauthorized, "Unauthorized"},
	ErrForbidden:    {ErrForbidden, http.StatusForbidden, "Forbidden"},
	ErrConflict:     {ErrConflict, http.StatusConflict, "Conflict"},
	ErrRateLimited:  {ErrRateLimited, http.StatusTooManyRequests, "Too many requests"},
	ErrUnavailable:  {ErrUnavailable, http.StatusServiceUnavailable, "Service unavailable"},

	ErrItemNotFound: {ErrItemNotFound, http.StatusNotFound, "Item not found"},
	ErrItemInvalid:  {ErrItemInvalid, http.StatusBadRequest, "Invalid item data"},
	ErrItemConflict: {ErrItemConflict, http.StatusConflict, "Item already exists"},
}

// Lookup returns the catalog entry for a code, or ErrInternal if unknown.
func Lookup(code ErrorCode) CatalogEntry {
	if e, ok := catalog[code]; ok {
		return e
	}
	return catalog[ErrInternal]
}

// AppError is an application error with a catalog code.
type AppError struct {
	Code    ErrorCode
	Message string
	Detail  string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// Kind returns the error kind for httpx mapping.
func (e *AppError) Kind() string {
	switch e.Code {
	case ErrNotFound, ErrItemNotFound:
		return "not_found"
	case ErrValidation, ErrItemInvalid:
		return "validation"
	case ErrUnauthorized:
		return "unauthorized"
	case ErrForbidden:
		return "forbidden"
	case ErrConflict, ErrItemConflict:
		return "conflict"
	case ErrRateLimited:
		return "rate_limited"
	case ErrUnavailable:
		return "unavailable"
	default:
		return "internal"
	}
}

// PublicDetail returns a safe message for the client.
func (e *AppError) PublicDetail() string {
	entry := Lookup(e.Code)
	if e.Detail != "" {
		return e.Detail
	}
	return entry.Message
}

// HTTPStatus returns the HTTP status for this error.
func (e *AppError) HTTPStatus() int {
	return Lookup(e.Code).Status
}

// New creates a new AppError.
func New(code ErrorCode, opts ...Option) *AppError {
	e := &AppError{Code: code, Message: Lookup(code).Message}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Option configures an AppError.
type Option func(*AppError)

// WithMessage sets a custom message.
func WithMessage(msg string) Option {
	return func(e *AppError) { e.Message = msg }
}

// WithDetail sets a public detail message.
func WithDetail(detail string) Option {
	return func(e *AppError) { e.Detail = detail }
}

// WithCause wraps an underlying error.
func WithCause(err error) Option {
	return func(e *AppError) { e.Cause = err }
}

// Domain-specific constructors
func NewItemNotFound(id string) *AppError {
	return New(ErrItemNotFound, WithDetail("Item not found: "+id))
}

func NewItemInvalid(detail string) *AppError {
	return New(ErrItemInvalid, WithDetail(detail))
}

func NewItemConflict(detail string) *AppError {
	return New(ErrItemConflict, WithDetail(detail))
}
