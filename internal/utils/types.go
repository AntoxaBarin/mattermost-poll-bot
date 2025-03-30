package utils

import (
	"net/url"

	"github.com/mattermost/mattermost/server/public/model"
)

type Config struct {
	URL      *url.URL
	BotToken string
	BotPort  string
	DBHost   string
	DBPort   string
}

type App struct {
	Config            Config
	MMClient          *model.Client4
	MMWebsocketClient *model.WebSocketClient
	MMUser            *model.User
	MMChannel         *model.Channel
}
