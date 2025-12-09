package i18n

import (
	"testing"
)

func TestNotificationMessagesHaveThresholdPlaceholder(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "English critical notification",
			message: MessagesForLanguage(LanguageEnglish).NotificationCritical,
		},
		{
			name:    "English warning notification",
			message: MessagesForLanguage(LanguageEnglish).NotificationWarning,
		},
		{
			name:    "French critical notification",
			message: MessagesForLanguage(LanguageFrench).NotificationCritical,
		},
		{
			name:    "French warning notification",
			message: MessagesForLanguage(LanguageFrench).NotificationWarning,
		},
		{
			name:    "Spanish critical notification",
			message: MessagesForLanguage(LanguageSpanish).NotificationCritical,
		},
		{
			name:    "Spanish warning notification",
			message: MessagesForLanguage(LanguageSpanish).NotificationWarning,
		},
		{
			name:    "German critical notification",
			message: MessagesForLanguage(LanguageGerman).NotificationCritical,
		},
		{
			name:    "German warning notification",
			message: MessagesForLanguage(LanguageGerman).NotificationWarning,
		},
		{
			name:    "Italian critical notification",
			message: MessagesForLanguage(LanguageItalian).NotificationCritical,
		},
		{
			name:    "Italian warning notification",
			message: MessagesForLanguage(LanguageItalian).NotificationWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !contains(tt.message, "{{threshold}}") {
				t.Errorf("expected message to contain {{threshold}} placeholder, got: %s", tt.message)
			}
			if !contains(tt.message, "{{count}}") {
				t.Errorf("expected message to contain {{count}} placeholder, got: %s", tt.message)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
