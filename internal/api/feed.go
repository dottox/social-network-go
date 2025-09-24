package api

import (
	"net/http"

	"github.com/dottox/social/internal/store"
)

// @Summary		Get user feed
// @Description	Get the feed for the authenticated user
// @Tags			feed
// @Accept			json
// @Produce		json
// @Param			limit	query		int		false	"Number of posts to return"	minimum(1)		maximum(25)	default(20)
// @Param			offset	query		int		false	"Number of posts to skip"	minimum(0)		default(0)
// @Param			sort	query		string	false	"Sort order: asc or desc"	enum(asc, desc)	default(desc)
// @Param			tags	query		string	false	"Comma-separated list of tags to filter by"
// @Param			search	query		string	false	"Search term to filter posts by title or content"
// @Param			since	query		string	false	"ISO 8601 date to filter posts created after this date"
// @Param			until	query		string	false	"ISO 8601 date to filter posts created before this date"
// @Success		200		{array}		model.Post
// @Failure		400		{object}	error
// @Failure		500		{object}	error
// @Security		BearerAuth
// @Router			/users/feed [get]
func (app *Application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
		Search: "",
		Since:  "",
		Until:  "",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Get the authenticated user from the context
	user := app.getAuthUserFromCtx(ctx)

	feed, err := app.Store.Posts.GetUserFeed(ctx, user.Id, fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
