package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dottox/social/internal/model"
	"github.com/dottox/social/internal/store"
	"github.com/go-chi/chi/v5"
)

// Handler to create a new comment
func (app *Application) createCommentHandler(w http.ResponseWriter, r *http.Request) {

	// Validate the postId param
	postIdParam := chi.URLParam(r, "postId")
	postId, err := strconv.ParseUint(postIdParam, 10, 32)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Took the payload from the request body
	// The payload will be a minimal comment model
	var payload model.CreateCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the struct of the payload
	// Required, max, etc
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Create the new comment if the payload had no errors
	comment := &model.Comment{
		Content: payload.Content,
		PostId:  uint32(postId),
		// TODO: Change this after auth
		UserId: 1,
	}

	// Get the request context
	ctx := r.Context()

	// Create the new comment in the repository
	// Basically inserting in the database
	// comment will be populated with the variable created at runtime: id, created_at
	if err := app.Store.Comments.Create(ctx, comment); err != nil {
		// We can switch here depending on the err to retrieve errors correctly
		app.internalServerError(w, r, err)
		return
	}

	// Write the response back to the user, with the http.StatusCreated.
	if err := writeJSON(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// Function to return a post by their Id
func (app *Application) getCommentsByPostHandler(w http.ResponseWriter, r *http.Request) {

	idParam := chi.URLParam(r, "postId")
	postId, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Get the request context
	ctx := r.Context()

	// Get the comments by their postId in the repository
	comments, err := app.Store.Comments.GetAllByPostId(ctx, uint32(postId))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrResourceNotFound):
			app.resourceNotFoundError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	// Write the post in JSON for the response
	if err := writeJSON(w, http.StatusOK, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
