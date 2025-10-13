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

	t.Run("should return 400 for bad request", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/users/NaN", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusBadRequest, rr.Code)
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

func TestFollowerUser(t *testing.T) {
	app := newTestApplication(t)
	mux := app.Mount()

	testToken, _ := app.Authenticator.GenerateToken(nil)

	t.Run("should not allow unauth requests", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/1/follow", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow auth requests", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/1/follow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNoContent, rr.Code)
	})

	t.Run("should return 400 for bad request", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/NaN/follow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should return 404 for non-existing user", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/0/follow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

	t.Run("should return 409 for existing following", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/409/follow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		_ = executeRequest(req, mux)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusConflict, rr.Code)
	})

}

func TestUnfollowerUser(t *testing.T) {
	app := newTestApplication(t)
	mux := app.Mount()

	testToken, _ := app.Authenticator.GenerateToken(nil)

	t.Run("should not allow unauth requests", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/1/unfollow", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow auth requests", func(t *testing.T) {
		// First follow the user
		req, err := http.NewRequest("PUT", "/v1/users/1/follow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		_ = executeRequest(req, mux)

		// Unfollow the user
		req, err = http.NewRequest("PUT", "/v1/users/1/unfollow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNoContent, rr.Code)
	})

	t.Run("should return 400 for bad request", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/NaN/unfollow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should return 404 for non-existing user", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/v1/users/0/unfollow", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNotFound, rr.Code)
	})

}
