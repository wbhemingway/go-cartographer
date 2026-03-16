package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/wbhemingway/go-cartographer/internal/models"
)

func (apiCfg *ApiConfig) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		key := strings.TrimPrefix(authHeader, "Bearer ")

		doc, err := apiCfg.firestoreClient.Collection("api_keys").Doc(key).Get(r.Context())
		if err != nil {
			slog.Warn("Invalid API key attempt", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var apiKey models.APIKey
		err = doc.DataTo(&apiKey)
		if err != nil || !apiKey.IsActive {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "UserID", apiKey.UserID)
		ctx = context.WithValue(ctx, "UserRole", apiKey.UserRole)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
