package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"poll_bot/internal/tt"
	"poll_bot/internal/utils"

	"github.com/mattermost/mattermost/server/public/model"
)

func sendMsgToChannel(app *utils.App, msg, channelID, replyToId string) {
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
	if strings.HasPrefix(post.Message, "_new_poll") {
		log.Printf("[Info]: Creating new poll: %s\n", post.Message)
		poll, err := createPoll(app, post)
		if err != nil {
			log.Printf("[Warning]: Failed to create poll: %s. Error: %s\n", post.Message, err)
			sendMsgToChannel(app, "Failed to create poll. Try again later.", post.ChannelId, post.RootId)
			return
		}
		log.Printf("[Info]: Poll (%s) successfully created\n", poll.Title)
		options := []string{}
		for k := range poll.Options {
			options = append(options, k)
		}

		msg := fmt.Sprintf("New poll: %s. ID: %d.\nOptions:\n%s", poll.Title, poll.ID, strings.Join(options, "\n")+"\n")
		sendMsgToChannel(app, msg, post.ChannelId, post.RootId)
	} else if strings.HasPrefix(post.Message, "_vote") {
		log.Printf("[Info]: Handle vote from user with ID %s: %s\n", post.UserId, post.Message)
		err := vote(app, post)
		if err != nil {
			log.Printf("[Warning]: Failed to vote: %s. Error: %s\n", post.Message, err)
			sendMsgToChannel(app, "Failed to vote. Try again later.", post.ChannelId, post.RootId)
			return
		}
		sendMsgToChannel(app, fmt.Sprintf("User with ID %s successfully voted", post.UserId), post.ChannelId, post.RootId)
	} else if strings.HasPrefix(post.Message, "_poll_res") {
		log.Printf("[Info]: Handle poll results from user with ID %s: %s\n", post.UserId, post.Message)
		poll, err := getPollResults(app, post.Message)
		if err != nil {
			log.Printf("[Warning]: Failed to get poll results: %s. Error: %s\n", post.Message, err)
			sendMsgToChannel(app, "Failed to get poll results. Try again later.", post.ChannelId, post.RootId)
			return
		}
		msg := fmt.Sprintf("Poll ID: %d results:\n", poll.ID)
		for k, v := range poll.Options {
			msg += fmt.Sprintf("%s: %d\n", k, len(v))
		}
		sendMsgToChannel(app, msg, post.ChannelId, post.RootId)
	} else if strings.HasPrefix(post.Message, "_cancel_poll") {
		log.Printf("[Info]: Handle poll calcellation from user with ID %s: %s\n", post.UserId, post.Message)
		poll, err := cancelPoll(app, post)
		if err != nil {
			log.Printf("[Warning]: Failed to cancel poll: %s. Error: %s\n", post.Message, err)
			sendMsgToChannel(app, "Failed to cancel poll. Try again later.", post.ChannelId, post.RootId)
			return
		}
		sendMsgToChannel(app, fmt.Sprintf("Poll ID: %d successfully canceled.", poll.ID), post.ChannelId, post.RootId)
	} else {
		log.Printf("[Info]: Ignore message: %s\n", post.Message)
	}
}

// Create poll: _new_poll <title>, <option 1>, <option 2>, ..., <option N>
func createPoll(app *utils.App, post *model.Post) (tt.Poll, error) {
	tokens := strings.Split(post.Message, ",")
	tokens[0] = strings.Join(strings.Split(tokens[0], " ")[1:], " ")
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}

	if len(tokens) < 3 {
		return tt.Poll{}, fmt.Errorf("bad request to create a poll. Usage:  _new_poll <title>, <option 1>, <option 2>, ..., <option N>")
	}
	if tokens[0] == "_new_poll" || tokens[0] == "_vote" || tokens[0] == "_poll_res" || tokens[0] == "_cancel_poll" {
		return tt.Poll{}, fmt.Errorf("poll's title cannot be bot's command: _new_poll, _vote, _poll_res, _cancel_poll")
	}
	options := map[string][]string{}
	for i := 1; i < len(tokens); i++ {
		options[tokens[i]] = []string{}
	}

	poll := tt.Poll{
		ID:        rand.Uint32(),
		Title:     tokens[0],
		Options:   options,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
		Creator:   post.UserId,
	}
	insertResult := tt.Insert(&app.Config, &poll)
	if !insertResult {
		return tt.Poll{}, fmt.Errorf("failed to insert Poll into Tarantool DB")
	}
	return poll, nil
}

// Vote: _vote <poll ID> <option>
func vote(app *utils.App, post *model.Post) error {
	tokens := strings.Split(post.Message, " ")
	if len(tokens) < 3 {
		return fmt.Errorf("bad vote request. Usage: _vote <poll ID> <option>")
	}
	option := strings.Join(tokens[2:], " ")
	ID, err := strconv.Atoi(tokens[1])
	if err != nil {
		return fmt.Errorf("invalid poll ID: %s", strconv.Itoa(ID))
	}
	poll, err := tt.Select(&app.Config, uint32(ID))
	if err != nil {
		return fmt.Errorf("failed to find poll")
	}

	if _, ok := poll.Options[option]; !ok {
		return fmt.Errorf("invalid option: %s", option)
	}
	for _, voters := range poll.Options {
		if slices.Contains(voters, post.UserId) {
			return fmt.Errorf("user %s already voted in this poll", post.UserId)
		}
	}
	updatedVoters := poll.Options[option]
	updatedVoters = append(updatedVoters, post.UserId)
	poll.Options[option] = updatedVoters

	err = tt.Update(&app.Config, &poll)
	if err != nil {
		return fmt.Errorf("failed to update poll")
	}
	return nil
}

// Get results: _poll_res <poll ID>
func getPollResults(app *utils.App, msg string) (tt.Poll, error) {
	tokens := strings.Split(msg, " ")
	if len(tokens) != 2 {
		return tt.Poll{}, fmt.Errorf("bad _poll_res request. Usage: _poll_res <poll ID>")
	}
	ID, err := strconv.Atoi(tokens[1])
	if err != nil {
		return tt.Poll{}, fmt.Errorf("invalid poll ID: %s", strconv.Itoa(ID))
	}
	poll, err := tt.Select(&app.Config, uint32(ID))
	if err != nil {
		return tt.Poll{}, fmt.Errorf("failed to find poll")
	}
	return poll, nil
}

// Cancel poll: _cancel_poll <poll ID>
func cancelPoll(app *utils.App, post *model.Post) (tt.Poll, error) {
	tokens := strings.Split(post.Message, " ")
	if len(tokens) != 2 {
		return tt.Poll{}, fmt.Errorf("bad _cancel_poll request. Usage: _cancel_poll <poll ID>")
	}
	ID, err := strconv.Atoi(tokens[1])
	if err != nil {
		return tt.Poll{}, fmt.Errorf("invalid poll ID: %s", strconv.Itoa(ID))
	}

	poll, err := tt.Select(&app.Config, uint32(ID))
	if err != nil {
		return tt.Poll{}, fmt.Errorf("failed to find poll")
	}
	if poll.Creator != post.UserId {
		return tt.Poll{}, fmt.Errorf("only creator can cancel poll")
	}

	err = tt.Delete(&app.Config, &poll)
	if err != nil {
		return tt.Poll{}, fmt.Errorf("failed to delete poll")
	}
	return poll, nil
}
