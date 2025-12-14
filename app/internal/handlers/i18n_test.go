package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"vcv/internal/handlers"
	"vcv/internal/i18n"

	"github.com/go-chi/chi/v5"
)

func TestRegisterI18nRoutes_Success(t *testing.T) {
	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/i18n", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify JSON response
	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Language == "" {
		t.Error("expected language to be set")
	}

	if len(response.Messages.AppTitle) == 0 {
		t.Error("expected messages to be populated")
	}
}

func TestRegisterI18nRoutes_WithQueryLang(t *testing.T) {
	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/i18n?lang=fr", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Language != i18n.LanguageFrench {
		t.Errorf("expected language %s, got %s", i18n.LanguageFrench, response.Language)
	}
}

func TestRegisterI18nRoutes_WithHXCurrentURL(t *testing.T) {
	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/i18n", nil)
	req.Header.Set("HX-Current-URL", "https://example.com?lang=es")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Language != i18n.LanguageSpanish {
		t.Errorf("expected language %s, got %s", i18n.LanguageSpanish, response.Language)
	}
}

func TestRegisterI18nRoutes_WithAcceptLanguage(t *testing.T) {
	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/i18n", nil)
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en;q=0.8")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should default to French based on Accept-Language
	if response.Language != i18n.LanguageFrench {
		t.Errorf("expected language %s, got %s", i18n.LanguageFrench, response.Language)
	}
}

func TestRegisterI18nRoutes_InvalidQueryLang(t *testing.T) {
	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/i18n?lang=invalid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should fall back to Accept-Language or default
	if response.Language == "" {
		t.Error("expected language to be set even with invalid query param")
	}
}

func TestRegisterI18nRoutes_InvalidHXCurrentURL(t *testing.T) {
	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/i18n", nil)
	req.Header.Set("HX-Current-URL", "not-a-valid-url")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Should handle invalid URL gracefully
	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}

func TestResolveLanguage_QueryParamPriority(t *testing.T) {
	// Test that query param takes priority over other methods
	req := httptest.NewRequest(http.MethodGet, "/api/i18n?lang=de", nil)
	req.Header.Set("HX-Current-URL", "https://example.com?lang=fr")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9")

	// We need to access the resolveLanguage function through the handler
	// Since it's not exported, we test it indirectly via the handler
	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Query param should win
	if response.Language != i18n.LanguageGerman {
		t.Errorf("expected language %s from query param, got %s", i18n.LanguageGerman, response.Language)
	}
}

func TestResolveLanguage_HXCurrentURLOverAcceptLanguage(t *testing.T) {
	// Test that HX-Current-URL takes priority over Accept-Language
	req := httptest.NewRequest(http.MethodGet, "/api/i18n", nil)
	req.Header.Set("HX-Current-URL", "https://example.com?lang=it")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")

	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// HX-Current-URL should win over Accept-Language
	if response.Language != i18n.LanguageItalian {
		t.Errorf("expected language %s from HX-Current-URL, got %s", i18n.LanguageItalian, response.Language)
	}
}

func TestResolveLanguage_ParsedURLWithInvalidLang(t *testing.T) {
	// Test when HX-Current-URL has an invalid language
	req := httptest.NewRequest(http.MethodGet, "/api/i18n", nil)
	req.Header.Set("HX-Current-URL", "https://example.com?lang=invalid")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9")

	router := chi.NewRouter()
	handlers.RegisterI18nRoutes(router)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var response i18n.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should fall back to Accept-Language
	if response.Language != i18n.LanguageGerman {
		t.Errorf("expected language %s from Accept-Language, got %s", i18n.LanguageGerman, response.Language)
	}
}
