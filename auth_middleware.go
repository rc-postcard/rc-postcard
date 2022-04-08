package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

const userContextKey = "user"

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

// authPersonalAccessToken will authenticate an Authorization header by
// forwarding a request to recurse.com API and cache a successful result
// in pacCache.
func authPersonalAccessToken(r *http.Request) (*User, error) {
	// get token
	pacToken := r.Header.Get("Authorization")
	if pacToken == "" {
		return nil, errors.New("PAT_NOT_FOUND")
	}
	// check cache
	if u, ok := pacCache[pacToken]; ok {
		return u, nil
	}
	// send request to recurse.com
	req, err := http.NewRequest(http.MethodGet, "https://recurse.com/api/v1/profiles/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", pacToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unauthorized")
	}

	// read body
	defer resp.Body.Close()
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// update cache
	pacCache[pacToken] = &user
	return pacCache[pacToken], nil
}
