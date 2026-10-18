package items_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/giovanibrioni/golang-template/internal/api"
	"github.com/giovanibrioni/golang-template/internal/core/pgtest"
	"github.com/giovanibrioni/golang-template/internal/items"
)

func TestItemsAPI(t *testing.T) {
	pool := pgtest.Require(t)
	handler := api.NewRouter(pgtest.RollbackBeginner(t, pool), items.NewService(items.NewPostgresRepository()))

	created := postItem(t, handler, `{"name":"widget"}`, http.StatusCreated)
	if created.Name != "widget" || created.ID == "" || created.CreatedAt.IsZero() {
		t.Fatalf("created = %+v", created)
	}

	rec := request(handler, http.MethodGet, "/api/v1/items", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d body = %s", rec.Code, rec.Body.String())
	}
	var listed items.ListItemsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("list json: %v", err)
	}
	if len(listed.Items) != 1 || listed.Items[0].ID != created.ID {
		t.Fatalf("list = %+v", listed)
	}

	rec = request(handler, http.MethodGet, "/api/v1/items/"+created.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d body = %s", rec.Code, rec.Body.String())
	}
	var fetched items.GetItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("get json: %v", err)
	}
	if fetched.Item.ID != created.ID || fetched.Item.Name != "widget" {
		t.Fatalf("fetched = %+v", fetched)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID")
	}

	rec = request(handler, http.MethodGet, "/api/v1/items/00000000-0000-0000-0000-000000000000", nil)
	assertDetail(t, rec, http.StatusNotFound, "Item not found")

	rec = request(handler, http.MethodPost, "/api/v1/items", strings.NewReader(`{"name":""}`))
	assertDetail(t, rec, http.StatusBadRequest, "Validation failed")
}

func postItem(t *testing.T, handler http.Handler, body string, want int) items.Item {
	t.Helper()
	rec := request(handler, http.MethodPost, "/api/v1/items", strings.NewReader(body))
	if rec.Code != want {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var created items.CreateItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("json: %v", err)
	}
	return created.Item
}

func request(handler http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func assertDetail(t *testing.T, rec *httptest.ResponseRecorder, status int, detail string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, status, rec.Body.String())
	}
	var body struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Detail != detail {
		t.Fatalf("detail = %q", body.Detail)
	}
}
