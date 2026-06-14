package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kimmykuang/selfskill/internal/plugin"
)

func TestMarketplaceAPI_ListEmpty(t *testing.T) {
	srv, _ := testServer(t)
	srv.Registry = plugin.NewRegistry()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/marketplaces", nil)
	srv.handleListMarketplaces(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestMarketplaceAPI_AddRequiresURL(t *testing.T) {
	srv, _ := testServer(t)
	srv.Registry = plugin.NewRegistry()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/marketplaces", bytes.NewReader([]byte(`{}`)))
	srv.handleAddMarketplace(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
