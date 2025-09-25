package api

import (
	"fmt"
	"net/http"
)

func (app *Application) logError(errorName string, r *http.Request, err error) {
	app.Logger.Errorw("error occurred", "error", err, "type", errorName, "method", r.Method, "url", r.URL.String())
}

func (app *Application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logError("internal server error", r, err)

	writeJSONError(w, http.StatusInternalServerError, "the server encountered an error")
}

func (app *Application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logError("bad request error", r, err)

	writeJSONError(w, http.StatusBadRequest, "bad request")
}

func (app *Application) resourceNotFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logError("resource not found", r, err)

	writeJSONError(w, http.StatusNotFound, "resource not found")
}

func (app *Application) resourceAlreadyExists(w http.ResponseWriter, r *http.Request, err error) {
	app.logError("resource already exists", r, err)

	writeJSONError(w, http.StatusConflict, "resource already exists")
}

func (app *Application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logError("unauthorized", r, err)

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	app.logError("unauthorized", r, err)

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) forbiddenError(w http.ResponseWriter, r *http.Request, err error) {
	app.logError("forbidden", r, err)

	writeJSONError(w, http.StatusForbidden, "forbidden")
}

func (app *Application) rateLimitExceededError(w http.ResponseWriter, r *http.Request, retryAfterText string) {
	app.logError("rate limit exceeded", r, fmt.Errorf("rate limit exceeded"))

	w.Header().Set("Retry-After", retryAfterText)

	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfterText)
}
