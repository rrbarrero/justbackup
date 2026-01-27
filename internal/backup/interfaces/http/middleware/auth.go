package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/rrbarrero/justbackup/internal/auth/application"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/auth"
	userApp "github.com/rrbarrero/justbackup/internal/user/application"
)

// UserIDKey is now imported from shared auth package

type AuthMiddleware struct {
	jwtService  *auth.JWTService
	authService *application.AuthService
	userService *userApp.UserService
}

func NewAuthMiddleware(jwtService *auth.JWTService, authService *application.AuthService, userService *userApp.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:  jwtService,
		authService: authService,
		userService: userService,
	}
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		authType := parts[0]
		credentials := parts[1]

		// 1. Basic Auth
		if authType == "Basic" {
			payload, err := base64.StdEncoding.DecodeString(credentials)
			if err != nil {
				http.Error(w, "Invalid basic auth header", http.StatusUnauthorized)
				return
			}
			pair := strings.SplitN(string(payload), ":", 2)
			if len(pair) != 2 {
				http.Error(w, "Invalid basic auth header", http.StatusUnauthorized)
				return
			}

			username := pair[0]
			password := pair[1]

			if m.userService != nil {
				user, err := m.userService.Authenticate(r.Context(), username, password)
				if err == nil && user != nil {
					// Basic auth successful
					// We could set UserIDKey here if we had the ID, but for now just allowing access
					// Actually, now we have the ID!
					ctx := context.WithValue(r.Context(), auth.UserIDKey, user.ID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// 2. Bearer Token (JWT or CLI Token)
		if authType == "Bearer" {
			// 2a. Try JWT
			claims, err := m.jwtService.ValidateToken(credentials)
			if err == nil {
				ctx := context.WithValue(r.Context(), auth.UserIDKey, claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2b. Try CLI Token (if service is available)
			if m.authService != nil {
				isValid, err := m.authService.ValidateToken(r.Context(), credentials)
				if err == nil && isValid {
					// CLI token is valid.
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
	})
}
