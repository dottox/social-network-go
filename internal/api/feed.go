package api

import (
	"net/http"

	"github.com/dottox/social/internal/store"
)

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

	// For now, we are using a hardcoded user ID.
	// TODO: Get this from the authenticated user context later.
	feed, err := app.Store.Posts.GetUserFeed(ctx, 2, fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
