package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dottox/social/internal/auth"
	"github.com/dottox/social/internal/store"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *Application {
	t.Helper()

	logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStore()
	mockAuthenticator := auth.NewMockAuthenticator()

	return &Application{
		Logger:        logger,
		Store:         *mockStore,
		Authenticator: mockAuthenticator,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected status %d; got %d", expected, actual)
	}
}
