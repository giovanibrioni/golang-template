package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/giovani/golang-template/internal/api"
	"github.com/giovani/golang-template/internal/core/pgtest"
	"github.com/giovani/golang-template/internal/items"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Run(m))
}

func TestHealth(t *testing.T) {
	pool := pgtest.Require(t)
	handler := api.NewRouter(pgtest.RollbackBeginner(t, pool), items.NewService(items.NewPostgresRepository()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status body = %+v", body)
	}
}
