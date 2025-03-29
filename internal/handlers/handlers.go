package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"poll_bot/internal/utils"

	"github.com/mattermost/mattermost/server/public/model"
)

func sendMsgToTalkingChannel(app *utils.App, msg, channelID, replyToId string) {
	post := &model.Post{}
	post.ChannelId = channelID
	post.Message = msg
	post.RootId = replyToId

	if _, _, err := app.MMClient.CreatePost(context.TODO(), post); err != nil {
		log.Println("[Warning]: Failed to create post")
	}
}

func SendAuthorizedRequest(url, endpoint, token string) (string, error) {
	req, err := http.NewRequest("GET", url+endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ListenToEvents(app *utils.App) {
	var err error
	failCount := 0
	for {
		app.MMWebsocketClient, err = model.NewWebSocketClient4(
			fmt.Sprintf("ws://%s", app.Config.URL.Host+app.Config.URL.Path),
			app.Config.BotToken,
		)
		if err != nil {
			failCount += 1
			if failCount > 100 {
				log.Fatal("[Error]: Can't connect to mattermost")
			}
			continue
		}
		log.Println("[Info]: Mattermost websocket connected. Listening...")

		app.MMWebsocketClient.Listen()

		for event := range app.MMWebsocketClient.EventChannel {
			go handleWebSocketEvent(app, event)
		}
	}
}

func handleWebSocketEvent(app *utils.App, event *model.WebSocketEvent) {
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		log.Println("[Info]: Could not cast event to *model.Post")
	}

	if post.UserId == app.MMUser.Id {
		return
	}
	handlePost(app, post)
}

func handlePost(app *utils.App, post *model.Post) {
	if strings.HasPrefix(post.Message, "_poll") {
		sendMsgToTalkingChannel(app, "Hello", post.ChannelId, post.RootId)
		log.Printf("[Info]: Handle message: %s\n", post.Message)
	} else {
		log.Printf("[Info]: Ignore message: %s\n", post.Message)
	}

}
