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

type userKey string

const userCtx userKey = "user"

// @Summary		Create a new user
// @Description	Create a new user with the given information
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			user	body		model.CreateUserPayload	true	"User payload"
// @Success		201		{object}	model.User
// @Failure		400		{object}	error
// @Failure		500		{object}	error
// @Security		ApiKeyAuth
// @Router			/users [post]
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

// @Summary		Get a user by ID
// @Description	Get a user by their ID
// @Tags			users
// @Produce		json
// @Param			userId	path		int	true	"User ID"
// @Success		200		{object}	model.User
// @Failure		400		{object}	error
// @Failure		404		{object}	error
// @Failure		500		{object}	error
// @Security		ApiKeyAuth
// @Router			/users/{userId} [get]
func (app *Application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserFromCtx(r.Context())

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Follow a user
// @Description	Follow a user by their ID
// @Tags			users
// @Param			userId	path	int	true	"User ID"
// @Success		204
// @Failure		400	{object}	error
// @Failure		404	{object}	error
// @Failure		409	{object}	error
// @Failure		500	{object}	error
// @Security		ApiKeyAuth
// @Router			/users/{userId}/follow [put]
func (app *Application) followUserHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	targetUser := getUserFromCtx(ctx)

	// TODO: get this users from auth later
	var followerUser uint32 = 1

	followAction := &model.FollowAction{
		TargetUserId: targetUser.Id,
		SenderUserId: followerUser,
	}

	err := app.Store.Followers.Follow(ctx, followAction)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrResourceNotFound):
			app.resourceNotFoundError(w, r, err)
			return
		case errors.Is(err, store.ErrResourceAlreadyExists):
			app.resourceAlreadyExists(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Unfollow a user
// @Description	Unfollow a user by their ID
// @Tags			users
// @Param			userId	path	int	true	"User ID"
// @Success		204
// @Failure		400	{object}	error
// @Failure		404	{object}	error
// @Failure		500	{object}	error
// @Security		ApiKeyAuth
// @Router			/users/{userId}/unfollow [put]
func (app *Application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	targetUser := getUserFromCtx(ctx)

	// TODO: get this users from auth later
	var unfollowerUser uint32 = 1

	followAction := &model.FollowAction{
		TargetUserId: targetUser.Id,
		SenderUserId: unfollowerUser,
	}

	err := app.Store.Followers.Unfollow(ctx, followAction)
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

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		userId, err := strconv.ParseUint(chi.URLParam(r, "userId"), 10, 32)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}

		user, err := app.Store.Users.GetById(ctx, uint32(userId))
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

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(ctx context.Context) *model.User {
	user, _ := ctx.Value(userCtx).(*model.User)
	return user
}
