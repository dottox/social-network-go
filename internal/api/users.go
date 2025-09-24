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

const userParamCtx userKey = "userParam"
const userAuthCtx userKey = "userAuth"

// @Summary		Get a user by ID
// @Description	Get a user by their ID
// @Tags			users
// @Produce		json
// @Param			userId	path		int	true	"User ID"
// @Success		200		{object}	model.User
// @Failure		400		{object}	error
// @Failure		404		{object}	error
// @Failure		500		{object}	error
// @Security		BearerAuth
// @Router			/users/{userId} [get]
func (app *Application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	user := app.getParamUserFromCtx(ctx)

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
// @Security		BearerAuth
// @Router			/users/{userId}/follow [put]
func (app *Application) followUserHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	targetUser := app.getParamUserFromCtx(ctx)
	followerUser := app.getAuthUserFromCtx(ctx)

	followAction := &model.FollowAction{
		TargetUserId: targetUser.Id,
		SenderUserId: followerUser.Id,
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
// @Security		BearerAuth
// @Router			/users/{userId}/unfollow [put]
func (app *Application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	targetUser := app.getParamUserFromCtx(ctx)
	unfollowerUser := app.getAuthUserFromCtx(ctx)

	followAction := &model.FollowAction{
		TargetUserId: targetUser.Id,
		SenderUserId: unfollowerUser.Id,
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

		ctx = context.WithValue(ctx, userParamCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) getParamUserFromCtx(ctx context.Context) *model.User {
	user, _ := ctx.Value(userParamCtx).(*model.User)
	return user
}

func (app *Application) getAuthUserFromCtx(ctx context.Context) *model.User {
	user, _ := ctx.Value(userAuthCtx).(*model.User)
	return user
}
