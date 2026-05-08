// Package admin provides routing for the Admin API endpoints.
package admin

import (
	"bytes"
	"context"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"prodlich/internal/application"
	"prodlich/internal/infrastructure/log"
)

type contextKey string

// AdminUserKey holds the authenticated *entity.AdminUser.
const AdminUserKey contextKey = "admin_user"

// CORSMiddleware applies a permissive CORS policy suitable for the admin UI.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics and logs stack traces.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := bytes.TrimSpace(debug.Stack())
				log.L().WithFields(map[string]any{
					"path":   r.URL.Path,
					"method": r.Method,
					"panic":  rec,
				}).Errorf("admin recovered from panic\n%s", stack)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// AuthConfig configures the unified auth middleware. JWT bearer tokens are
// always accepted; if APIToken is non-nil, "Authorization: Token <secret>"
// (and the convenience "Bearer <prefix>.<secret>" form) is also accepted and
// resolved into the owning admin user.
type AuthConfig struct {
	Auth      *application.AdminAuthUseCase
	APIToken  *application.APIToken
	Admins    *application.User
	OpenPaths []string
}

// AuthMiddleware accepts either a JWT bearer token or an API token in the
// "Authorization: Token <secret>" form.
func AuthMiddleware(cfg AuthConfig) func(http.Handler) http.Handler {
	open := make(map[string]struct{}, len(cfg.OpenPaths))
	for _, p := range cfg.OpenPaths {
		open[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := open[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}
			scheme := strings.ToLower(parts[0])
			secret := parts[1]
			switch scheme {
			case "bearer":
				if admin, err := cfg.Auth.ValidateToken(r.Context(), secret); err == nil {
					ctx := context.WithValue(r.Context(), AdminUserKey, admin)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// Fall through to token auth if the bearer payload looks like
				// an API token (prefix.secret) and the token UC is wired.
				if cfg.APIToken != nil && strings.Contains(secret, ".") {
					if admin, ok := authenticateAPIToken(r, cfg, secret); ok {
						ctx := context.WithValue(r.Context(), AdminUserKey, admin)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			case "token":
				if cfg.APIToken == nil {
					http.Error(w, "api tokens disabled", http.StatusUnauthorized)
					return
				}
				if admin, ok := authenticateAPIToken(r, cfg, secret); ok {
					ctx := context.WithValue(r.Context(), AdminUserKey, admin)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			default:
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			}
		})
	}
}

// authenticateAPIToken validates a programmatic token and returns the admin
// user that owns it.
func authenticateAPIToken(r *http.Request, cfg AuthConfig, secret string) (any, bool) {
	tok, err := cfg.APIToken.Authenticate(r.Context(), secret)
	if err != nil {
		return nil, false
	}
	if cfg.Admins == nil {
		return nil, false
	}
	user, err := cfg.Admins.Get(r.Context(), tok.UserID)
	if err != nil {
		return nil, false
	}
	return user, true
}

// JWTMiddleware retains the legacy JWT-only behaviour for callers that don't
// have an API token use case wired yet.
func JWTMiddleware(authUC *application.AdminAuthUseCase, openPaths ...string) func(next http.Handler) http.Handler {
	return AuthMiddleware(AuthConfig{Auth: authUC, OpenPaths: openPaths})
}

// LoggingMiddleware emits structured logs for every admin request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := acquireRW(w)
		defer releaseRW(ww)
		next.ServeHTTP(ww, r)
		log.L().WithFields(map[string]any{
			"component":   "admin",
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      ww.statusCode,
			"duration_ms": time.Since(start).Milliseconds(),
			"remote_addr": r.RemoteAddr,
		}).Info("admin request")
	})
}

// pooled response writer wrapper for the admin server.
var rwPool = sync.Pool{New: func() any { return &responseWriterWrapper{} }}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func acquireRW(w http.ResponseWriter) *responseWriterWrapper {
	rw := rwPool.Get().(*responseWriterWrapper)
	rw.ResponseWriter = w
	rw.statusCode = http.StatusOK
	rw.wroteHeader = false
	return rw
}

func releaseRW(rw *responseWriterWrapper) {
	rw.ResponseWriter = nil
	rwPool.Put(rw)
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(b)
}
