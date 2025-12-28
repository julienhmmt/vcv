package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		code     string
		expected Language
		ok       bool
	}{
		{"en", LanguageEnglish, true},
		{"EN", LanguageEnglish, true},
		{"  en  ", LanguageEnglish, true},
		{"fr", LanguageFrench, true},
		{"es", LanguageSpanish, true},
		{"de", LanguageGerman, true},
		{"it", LanguageItalian, true},
		{"unknown", "", false},
		{"", "", false},
		{"pt", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got, ok := GetLanguage(tt.code)
			if ok != tt.ok || got != tt.expected {
				t.Errorf("GetLanguage(%q) = (%v, %v), want (%v, %v)", tt.code, got, ok, tt.expected, tt.ok)
			}
		})
	}
}

func TestFromAcceptLanguage_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected Language
	}{
		{"empty", "", LanguageEnglish},
		{"single language", "fr", LanguageFrench},
		{"language with region", "fr-FR", LanguageFrench},
		{"multiple languages with quality", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7", LanguageFrench},
		{"unknown language falls back", "pt-BR,pt;q=0.9", LanguageEnglish},
		{"mixed case", "FR-fr", LanguageFrench},
		{"whitespace", "  es  ", LanguageSpanish},
		{"complex header", "de-DE,de;q=0.9,en;q=0.8,fr;q=0.7", LanguageGerman},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromAcceptLanguage(tt.header)
			if got != tt.expected {
				t.Errorf("FromAcceptLanguage(%q) = %v, want %v", tt.header, got, tt.expected)
			}
		})
	}
}

func TestResolveLanguage_QueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?lang=fr", nil)
	got := ResolveLanguage(req)
	if got != LanguageFrench {
		t.Errorf("ResolveLanguage with query param = %v, want %v", got, LanguageFrench)
	}
}

func TestResolveLanguage_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "es"})
	got := ResolveLanguage(req)
	if got != LanguageSpanish {
		t.Errorf("ResolveLanguage with cookie = %v, want %v", got, LanguageSpanish)
	}
}

func TestResolveLanguage_HXCurrentURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("hx-current-url", "https://example.com?lang=de")
	got := ResolveLanguage(req)
	if got != LanguageGerman {
		t.Errorf("ResolveLanguage with HX-Current-URL = %v, want %v", got, LanguageGerman)
	}
}

func TestResolveLanguage_AcceptLanguage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "it-IT,it;q=0.9")
	got := ResolveLanguage(req)
	if got != LanguageItalian {
		t.Errorf("ResolveLanguage with Accept-Language = %v, want %v", got, LanguageItalian)
	}
}

func TestResolveLanguage_Priority(t *testing.T) {
	// Query param should win over cookie
	req := httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	req.Header.Set("Accept-Language", "es")
	got := ResolveLanguage(req)
	if got != LanguageEnglish {
		t.Errorf("ResolveLanguage priority = %v, want %v (query param should win)", got, LanguageEnglish)
	}
}

func TestResolveLanguage_HXOverCookie(t *testing.T) {
	// HX-Current-URL should win over cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("hx-current-url", "https://example.com?lang=de")
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	req.Header.Set("Accept-Language", "es")
	got := ResolveLanguage(req)
	if got != LanguageGerman {
		t.Errorf("ResolveLanguage HX priority = %v, want %v (HX should win over cookie)", got, LanguageGerman)
	}
}

func TestResolveLanguage_InvalidHXURL(t *testing.T) {
	// Invalid HX-Current-URL should fall back to cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("hx-current-url", "://invalid-url")
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	got := ResolveLanguage(req)
	if got != LanguageFrench {
		t.Errorf("ResolveLanguage with invalid HX URL = %v, want %v (should fall back to cookie)", got, LanguageFrench)
	}
}

func TestResolveLanguage_Default(t *testing.T) {
	// No language hints should default to English
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := ResolveLanguage(req)
	if got != LanguageEnglish {
		t.Errorf("ResolveLanguage default = %v, want %v", got, LanguageEnglish)
	}
}

func TestMessagesForLanguage_AllLanguages(t *testing.T) {
	languages := []Language{LanguageEnglish, LanguageFrench, LanguageSpanish, LanguageGerman, LanguageItalian}
	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			msg := MessagesForLanguage(lang)
			if msg.AppTitle == "" {
				t.Errorf("MessagesForLanguage(%v).AppTitle is empty", lang)
			}
			if msg.ButtonClose == "" {
				t.Errorf("MessagesForLanguage(%v).ButtonClose is empty", lang)
			}
			if msg.StatusLabelValid == "" {
				t.Errorf("MessagesForLanguage(%v).StatusLabelValid is empty", lang)
			}
		})
	}
}

func TestTranslations_MapExists(t *testing.T) {
	if len(Translations) != 5 {
		t.Errorf("Translations map has %d entries, want 5", len(Translations))
	}
	for lang := range Translations {
		if _, ok := GetLanguage(lang); !ok {
			t.Errorf("Translations contains invalid language code: %s", lang)
		}
	}
}
