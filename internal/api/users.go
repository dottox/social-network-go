package api

import (
	"net/http"

	"github.com/dottox/social/internal/model"
)

// Handler to create a new user
func (app *Application) createUserHandler(w http.ResponseWriter, r *http.Request) {

	// Took the payload from the request body
	// The payload will be a minimal User model
	var payload model.CreateUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Create the new user if the payload had no errors
	user := &model.User{
		Username: payload.Username,
		Password: payload.Password,
		Email:    payload.Email,
	}

	// Get the request context
	ctx := r.Context()

	// Create the new user in the repository
	// Basically inserting it in the database
	// user will be populated with the variable created at runtime: id & created_at
	if err := app.Store.Users.Create(ctx, user); err != nil {
		// We can switch here depending on the err to retrieve errors correctly
		app.internalServerError(w, r, err)
		return
	}

	// Write the response back to the user, with the http.StatusCreated.
	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
