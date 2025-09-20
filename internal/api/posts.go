package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/dottox/social/internal/model"
	"github.com/dottox/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

// Handler to create a new post
func (app *Application) createPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get the request context
	ctx := r.Context()

	// Took the payload from the request body
	// The payload will be a minimal Post model
	var payload model.CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the struct of the payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Check if tags is empty, if so assign an empty array
	// Tags are optional
	if payload.Tags == nil {
		payload.Tags = []string{}
	}

	// Create the new post if the payload had no errors
	post := &model.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		// Change after auth
		UserId: 1,
	}

	// Create the new post in the repository
	// Basically inserting in the database
	// post will be populated with the variable created at runtime: id, created_at, updated_at
	if err := app.Store.Posts.Create(ctx, post); err != nil {
		// We can switch here depending on the err to retrieve errors correctly
		app.internalServerError(w, r, err)
		return
	}

	// Write the response back to the user, with the http.StatusCreated.
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get the post from the context
	ctx := r.Context()
	post := getPostFromCtx(ctx)

	// Write the post in JSON for the response
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// Handler to update a post by their Id
func (app *Application) updatePostHandler(w http.ResponseWriter, r *http.Request) {

	// Get the request context
	ctx := r.Context()
	post := getPostFromCtx(ctx)

	// Took the payload from the request body
	// The payload will be a minimal Post model
	var payload model.UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Check if at least one field is provided
	if (payload.Title == nil) && (payload.Content == nil) {
		app.badRequestError(w, r, errors.New("at least one field must be provided to update the post"))
		return
	}

	// Update the post fields if they are provided in the payload
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	// Update the Post by Id in the repository
	err := app.Store.Posts.Update(ctx, post)
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
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) deletePostHandler(w http.ResponseWriter, r *http.Request) {

	// Get the request context
	ctx := r.Context()
	post := getPostFromCtx(ctx)

	// Delete the Post by Id in the repository
	err := app.Store.Posts.DeleteById(ctx, post.Id)
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
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get the request context
		ctx := r.Context()

		idParam := chi.URLParam(r, "postId")
		id, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}

		// Get the Post by Id in the repository
		post, err := app.Store.Posts.GetById(ctx, uint32(id))
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

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(ctx context.Context) *model.Post {
	post, _ := ctx.Value(postCtx).(*model.Post)
	return post
}
