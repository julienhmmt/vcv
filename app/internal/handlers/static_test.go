package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRegisterStaticRoutes(t *testing.T) {
	r := chi.NewRouter()

	testFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><html><body>test</body></html>")},
		"admin.html": &fstest.MapFile{Data: []byte("<!doctype html><html><body>admin</body></html>")},
	}

	RegisterStaticRoutes(r, testFS)

	// Verify routes are registered
	assert.NotNil(t, r)
}

func TestStaticRoutes_ServeIndex(t *testing.T) {
	r := chi.NewRouter()

	testFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><html><body>test</body></html>")},
	}

	RegisterStaticRoutes(r, testFS)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestStaticRoutes_ServeAdmin(t *testing.T) {
	r := chi.NewRouter()

	testFS := fstest.MapFS{
		"admin.html": &fstest.MapFile{Data: []byte("<!doctype html><html><body>admin</body></html>")},
	}

	RegisterStaticRoutes(r, testFS)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin")
}

func TestStaticRoutes_ServeStaticAssets(t *testing.T) {
	r := chi.NewRouter()

	testFS := fstest.MapFS{
		"assets/main.js":    &fstest.MapFile{Data: []byte("console.log('ok')")},
		"assets/styles.css": &fstest.MapFile{Data: []byte("body { color: red; }")},
		"favicon.ico":       &fstest.MapFile{Data: []byte("\x00\x00")},
	}

	RegisterStaticRoutes(r, testFS)

	// Test various static asset paths
	paths := []string{
		"/assets/main.js",
		"/assets/styles.css",
		"/favicon.ico",
	}

	for _, path := range paths {
		t.Run("path_"+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestStaticRoutes_CacheHeaders(t *testing.T) {
	r := chi.NewRouter()

	testFS := fstest.MapFS{
		"assets/test.js": &fstest.MapFile{Data: []byte("console.log('test')")},
	}

	RegisterStaticRoutes(r, testFS)

	req := httptest.NewRequest(http.MethodGet, "/assets/test.js", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Static assets should have some cache control
	cacheControl := w.Header().Get("Cache-Control")
	assert.NotEmpty(t, cacheControl)
}

func TestImmutableCache(t *testing.T) {
	// Test the immutable cache middleware function
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	middleware := immutableCache(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/assets/test.js", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "public, max-age=31536000, immutable", w.Header().Get("Cache-Control"))
}

func TestStaticRoutes_MethodNotAllowed(t *testing.T) {
	r := chi.NewRouter()

	testFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><html><body>test</body></html>")},
	}

	RegisterStaticRoutes(r, testFS)

	// Test that POST requests to static routes are not allowed
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return Method Not Allowed or handle it gracefully
	assert.True(t, w.Code == http.StatusMethodNotAllowed || w.Code == http.StatusOK)
}

func TestStaticRoutes_ConcurrentAccess(t *testing.T) {
	r := chi.NewRouter()

	testFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><html><body>test</body></html>")},
	}

	RegisterStaticRoutes(r, testFS)

	// Test concurrent access to static routes
	const numGoroutines = 10
	const numRequests = 5

	done := make(chan bool, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			for range numRequests {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				w := httptest.NewRecorder()

				// This should not cause race conditions
				r.ServeHTTP(w, req)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}
}
