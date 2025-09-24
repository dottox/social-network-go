package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dottox/social/internal/model"
	"github.com/dottox/social/web/components"
)

func (wApp *WebApp) FeedHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Fetching feed from API:", wApp.apiUrl+"/v1/users/feed")
	feedJson, err := http.Get(wApp.apiUrl + "/v1/users/feed")
	if err != nil {
		http.Error(w, "Failed to fetch feed", http.StatusInternalServerError)
		return
	}
	defer feedJson.Body.Close()

	// Parse the JSON
	var data struct {
		Posts []*model.Post `json:"data"`
	}

	if err := json.NewDecoder(feedJson.Body).Decode(&data); err != nil {
		http.Error(w, "Failed to parse feed", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Fetched posts: %+v\n", data.Posts)

	templ.Handler(components.Feed(data.Posts)).ServeHTTP(w, r)
}
