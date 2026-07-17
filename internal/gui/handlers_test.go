package gui

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestGuardLocalAPI_LoopbackHostNoOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Host = "127.0.0.1:9999"
	rec := httptest.NewRecorder()

	guardLocalAPI(okHandler)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGuardLocalAPI_NonLoopbackHostRejected(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Host = "evil.example.com"
	rec := httptest.NewRecorder()

	guardLocalAPI(okHandler)(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestGuardLocalAPI_CrossSiteOriginRejected(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Host = "127.0.0.1:9999"
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()

	guardLocalAPI(okHandler)(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestGuardLocalAPI_LoopbackOriginAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Host = "127.0.0.1:9999"
	req.Header.Set("Origin", "http://127.0.0.1:9999")
	rec := httptest.NewRecorder()

	guardLocalAPI(okHandler)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGuardLocalAPI_LocalhostHostAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Host = "localhost:9999"
	rec := httptest.NewRecorder()

	guardLocalAPI(okHandler)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHandler_APIRoutesWrapped verifies the four /api/* routes registered in
// Handler are actually wrapped with the guard by exercising them through the
// real mux with a cross-origin Host.
func TestHandler_APIRoutesWrapped(t *testing.T) {
	storage := NewStorage(10)
	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	handler := Handler(staticFS, storage)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/ingest"},
		{http.MethodGet, "/api/events"},
		{http.MethodGet, "/api/stream"},
		{http.MethodPost, "/api/events/clear"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Host = "evil.example.com"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s %s: status = %d, want %d", tc.method, tc.path, rec.Code, http.StatusForbidden)
		}
	}
}
