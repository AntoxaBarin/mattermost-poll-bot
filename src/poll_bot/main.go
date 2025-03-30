package main

import (
	"context"
	"log"

	"poll_bot/internal/handlers"
	"poll_bot/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mattermost/mattermost/server/public/model"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	app := utils.CreateApp()
	app.MMClient = model.NewAPIv4Client(app.Config.URL.String())
	app.MMClient.SetToken(app.Config.BotToken)

	if user, _, err := app.MMClient.GetUser(context.TODO(), "me", ""); err != nil {
		log.Fatal("[Error]: Failed to login bot into Mattermost")
	} else {
		log.Println("[Info]: Successfully logged into Mattermost")
		app.MMUser = user
		log.Printf("[Info]: User info. FirstName: %s, email: %s", user.FirstName, user.Email)
	}

	handlers.ListenToEvents(app)
}
