package api

import (
	"net/http"
)

// Handler to check if you can reach the API
//
//	@Summary		Health Check
//	@Description	Simple health check to see if the API is running
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.Config.Env,
		"version": app.Config.Version,
	}

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}
}
