package main

import (
	"context"
	"log"
	"net/http"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasValidAuth := false
		user, err := authPersonalAccessToken(r)

		if err == nil {
			hasValidAuth = true
		} else if err.Error() == "PAT_NOT_FOUND" {
			currentSession, err := getSession(r)
			if err == nil && currentSession.isAuthenticated() {
				user = &currentSession.User
				hasValidAuth = true
			}
		}

		if hasValidAuth {
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	})
}
