package middleware

import (
	"strings"
)

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return strings.TrimSpace(header)
}
