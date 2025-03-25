package main

import (
	"log"
	"net/http"

	"poll_bot/internal/hadlers"
	"poll_bot/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	utils.InitEnv()
	botToken, err := utils.GetEnvWrapper("BOT_TOKEN")
	if err != nil {
		log.Fatal(err)
	}
	botPort, err := utils.GetEnvWrapper("BOT_PORT")
	if err != nil {
		log.Fatal(err)
	}
	mattermostURL, err := utils.GetEnvWrapper("MM_URL")
	if err != nil {
		log.Fatal(err)
	}

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		resp, err := hadlers.SendAuthorizedRequest(mattermostURL, "/users", botToken)
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte(resp))
	})

	http.ListenAndServe(":"+botPort, r)
}
