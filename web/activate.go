package web

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/dottox/social/web/components"
)

func (wApp *WebApp) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {

	token := r.URL.Query().Get("token")
	if token == "" {
		templ.Handler(components.UserActivateError("Invalid activation token")).ServeHTTP(w, r)
		return
	}

	// Create a new request
	client := &http.Client{}
	req, err := http.NewRequest("PUT", wApp.apiUrl+"/v1/auth/user/activate?token="+token, nil)
	if err != nil {
		templ.Handler(components.UserActivateError("Failed to activate user")).ServeHTTP(w, r)
		return
	}

	// Sent the request
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		templ.Handler(components.UserActivateError("Failed to activate user")).ServeHTTP(w, r)
		return
	}

	templ.Handler(components.UserActivate()).ServeHTTP(w, r)
}
