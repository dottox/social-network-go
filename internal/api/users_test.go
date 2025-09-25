package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestGetUser(t *testing.T) {
	app := newTestApplication(t)
	mux := app.Mount()

	testToken, _ := app.Authenticator.GenerateToken(nil)

	t.Run("should not allow unauth requests", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow auth requests", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 404 for non-existing user", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/users/0", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

	t.Run("should return user x when calling with user x id", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		// Check if "id":1 is in the response body
		expected := `"id":1`
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("expected response body to contain %s; got %s", expected, rr.Body.String())
		}

	})
}
