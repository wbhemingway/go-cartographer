package main

import (
	"net/http"
	"os"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("CARTOGRAPHER_API_KEY")

		if apiKey != "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer "+apiKey {
				http.Error(w, "Unauthorized: Invalid or missing API Key", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	}
}
