package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"vcv/internal/logger"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// RequestIDKey is the context key for request ID.
const RequestIDKey contextKey = "request_id"

// Logger logs HTTP requests with timing information using zerolog.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)
		requestID := GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, wrapped.statusCode, float64(duration.Milliseconds())).
			Str("request_id", requestID).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Msg("HTTP request")
	})
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Recoverer recovers from panics and returns a 500 error.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())
				logger.PanicEvent(err, string(debug.Stack())).
					Str("request_id", requestID).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("Panic recovered")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns a CORS middleware with the given configuration.
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			allowed := false
			for _, o := range config.AllowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}
			if !allowed {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods))
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", itoa(config.MaxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID adds a unique request ID to each request and stores it in context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := sanitizeRequestID(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = generateRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// joinStrings joins strings with comma separator.
func joinStrings(s []string) string {
	if len(s) == 0 {
		return ""
	}
	result := s[0]
	for i := 1; i < len(s); i++ {
		result += ", " + s[i]
	}
	return result
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func generateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return itoa(int(time.Now().UnixNano() % 1000000000))
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

func sanitizeRequestID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) > 128 {
		return ""
	}
	for _, r := range trimmed {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		switch r {
		case '-', '_', '.', ':':
			continue
		default:
			return ""
		}
	}
	return trimmed
}

func CSRFProtection(next http.Handler) http.Handler {
	safeMethods := map[string]struct{}{http.MethodGet: {}, http.MethodHead: {}, http.MethodOptions: {}}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := safeMethods[r.Method]; ok {
			next.ServeHTTP(w, r)
			return
		}
		fetchSite := strings.ToLower(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")))
		if fetchSite == "cross-site" || fetchSite == "same-site" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" {
			if sameOrigin(origin, targetOrigin(r)) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		referer := strings.TrimSpace(r.Header.Get("Referer"))
		if referer != "" {
			parsed, err := url.Parse(referer)
			if err == nil {
				refererOrigin := parsed.Scheme + "://" + parsed.Host
				if sameOrigin(refererOrigin, targetOrigin(r)) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func targetOrigin(r *http.Request) string {
	proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	return proto + "://" + host
}

func sameOrigin(left string, right string) bool {
	return strings.EqualFold(strings.TrimSuffix(left, "/"), strings.TrimSuffix(right, "/"))
}

func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}
			if r.ContentLength > maxBytes {
				http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

type RateLimitConfig struct {
	MaxRequests        int
	Window             time.Duration
	MaxEntries         int
	ExemptPaths        []string
	ExemptPathPrefixes []string
}

type rateLimiterEntry struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	config  RateLimitConfig
	entries map[string]rateLimiterEntry
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{MaxRequests: 300, Window: 1 * time.Minute, MaxEntries: 10_000}
}

func RateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	limiter := &rateLimiter{config: config, entries: make(map[string]rateLimiterEntry)}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if shouldSkipRateLimit(r, config) {
				next.ServeHTTP(w, r)
				return
			}
			allowed, retryAfter := limiter.allow(time.Now(), clientIP(r))
			if !allowed {
				if retryAfter > 0 {
					w.Header().Set("Retry-After", itoa(retryAfter))
				}
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func shouldSkipRateLimit(r *http.Request, config RateLimitConfig) bool {
	path := r.URL.Path
	for _, exempt := range config.ExemptPaths {
		if path == exempt {
			return true
		}
	}
	for _, prefix := range config.ExemptPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func (l *rateLimiter) allow(now time.Time, key string) (bool, int) {
	if key == "" {
		return true, 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prune(now)
	entry := l.entries[key]
	if entry.resetAt.IsZero() || now.After(entry.resetAt) {
		entry = rateLimiterEntry{count: 0, resetAt: now.Add(l.config.Window)}
	}
	entry.count++
	l.entries[key] = entry
	if entry.count <= l.config.MaxRequests {
		return true, 0
	}
	retryAfterSeconds := int(time.Until(entry.resetAt).Seconds())
	if retryAfterSeconds < 0 {
		retryAfterSeconds = 0
	}
	return false, retryAfterSeconds
}

func (l *rateLimiter) prune(now time.Time) {
	for key, entry := range l.entries {
		if now.After(entry.resetAt) {
			delete(l.entries, key)
		}
	}
	if l.config.MaxEntries <= 0 {
		return
	}
	if len(l.entries) <= l.config.MaxEntries {
		return
	}
	for len(l.entries) > l.config.MaxEntries {
		var oldestKey string
		oldestResetAt := now.Add(365 * 24 * time.Hour)
		for key, entry := range l.entries {
			if entry.resetAt.Before(oldestResetAt) {
				oldestKey = key
				oldestResetAt = entry.resetAt
			}
		}
		if oldestKey == "" {
			break
		}
		delete(l.entries, oldestKey)
	}
}

func clientIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			value := strings.TrimSpace(parts[0])
			if value != "" {
				return value
			}
		}
	}
	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

// SecurityHeaders adds security-related HTTP headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=()")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		if r.TLS != nil || strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}
