package main

import (
	"fmt"
	"net/http"
	"testing"
)

// go test
func TestGetUser(t *testing.T) {
	app := NewTestApplication(t)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatalf("could not generate test token: %v", err)
	}

	t.Run("should not allow unauthenticated user to fetch user profile", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/10", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}
		rr := executeRequest(req, mux)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})
	t.Run("should allow authenticated user to fetch user profile", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/10", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))
		rr := executeRequest(req, mux)
		checkResponseCode(t, http.StatusOK, rr.Code)
	})
}
