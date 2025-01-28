package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Shadowcyng/goSocial/internal/auth"
	"github.com/Shadowcyng/goSocial/internal/store"
	"github.com/Shadowcyng/goSocial/internal/store/cache"
	"go.uber.org/zap"
)

func NewTestApplication(t *testing.T) *application {
	t.Helper()
	// logger := zap.NewNop().Sugar()
	logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStore()
	cacheMockStore := cache.NewMockCache()
	testAuth := &auth.TestAuthenticator{}
	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  cacheMockStore,
		authenticator: testAuth,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected status code to be %d, got %d", expected, actual)
	}
}
