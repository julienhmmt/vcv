package i18n

import "testing"

func TestMessagesForLanguage(t *testing.T) {
	if MessagesForLanguage(LanguageEnglish).AppTitle == "" {
		t.Fatalf("expected messages for English")
	}
	if MessagesForLanguage(LanguageFrench).AppTitle == "" {
		t.Fatalf("expected messages for French")
	}
	unknown := MessagesForLanguage(Language("xx"))
	if unknown.AppTitle != englishMessages.AppTitle {
		t.Fatalf("expected fallback to English")
	}
}

func TestFromAcceptLanguage(t *testing.T) {
	tests := []struct {
		header   string
		expected Language
	}{
		{"fr-FR,fr;q=0.9", LanguageFrench},
		{"es;q=0.8", LanguageSpanish},
		{"de,en;q=0.5", LanguageGerman},
		{"it-IT,it;q=0.9", LanguageItalian},
		{"", LanguageEnglish},
		{"pt-BR, en-US", LanguageEnglish},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := FromAcceptLanguage(tt.header)
			if got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
