package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"poll_bot/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mattermost/mattermost/server/public/model"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	cfg := utils.CreateConfig()
	mmClient := model.NewAPIv4Client(cfg.URL)
	mmClient.SetToken(cfg.BotToken)

	var perPage int = 10
	var page int
	for {
		users, _, err := mmClient.GetUsers(context.TODO(), page, perPage, "")
		if err != nil {
			log.Printf("error fetching users: %v", err)
			return
		}

		for _, u := range users {
			fmt.Printf("%s\n", u.Username)
		}

		if len(users) < perPage {
			break
		}

		page++
	}

	http.ListenAndServe(":"+cfg.BotPort, r)
}
