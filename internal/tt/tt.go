package tt

import (
	"context"
	"fmt"
	"log"
	"poll_bot/internal/utils"
	"time"

	"github.com/tarantool/go-tarantool/v2"
	_ "github.com/tarantool/go-tarantool/v2/datetime"
	_ "github.com/tarantool/go-tarantool/v2/decimal"
	_ "github.com/tarantool/go-tarantool/v2/uuid"
)

const (
	spaceName = "polls"
	user      = "guest"
)

type Poll struct {
	ID        uint32
	Title     string
	Options   map[string][]string
	Timestamp string
	Creator   string
}

func Insert(config *utils.Config, poll *Poll) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address: config.DBHost + ":" + config.DBPort,
		User:    user,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Println("[Tarantool]: Connection refused:", err)
		return false
	}

	data, err := conn.Do(
		tarantool.NewInsertRequest(spaceName).Tuple([]interface{}{
			poll.ID,
			poll.Title,
			poll.Options,
			poll.Timestamp,
			poll.Creator,
		})).Get()
	if err != nil {
		log.Println("[Tarantool]: Error:", err)
		return false
	} else {
		log.Println("[Tarantool]: Inserted data:", data)
		return true
	}
}

func Select(config *utils.Config, ID uint32) (Poll, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address: config.DBHost + ":" + config.DBPort,
		User:    user,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Println("[Tarantool]: Connection refused:", err)
		return Poll{}, fmt.Errorf("[Tarantool]: failed to connect to Tarantool")
	}

	polls := []Poll{}
	selectReq := tarantool.NewSelectRequest(spaceName).
		Index(0).
		Limit(1).
		Iterator(tarantool.IterEq).
		Key([]any{ID})
	err = conn.Do(selectReq).GetTyped(&polls)
	if err != nil {
		log.Printf("[Tarantool]: Failed to Select: %s", err.Error())
		return Poll{}, fmt.Errorf("[Tarantool]: failed to Select")
	} else if len(polls) == 0 {
		log.Printf("[Tarantool]: No poll with ID: %d", ID)
		return Poll{}, fmt.Errorf("[Tarantool]: No poll with ID: %d", ID)
	}
	return polls[0], nil
}

func Update(config *utils.Config, poll *Poll) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address: config.DBHost + ":" + config.DBPort,
		User:    user,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Println("[Tarantool]: Connection refused:", err)
		return fmt.Errorf("[Tarantool]: failed to connect to Tarantool")
	}

	_, err = conn.Do(
		tarantool.NewUpdateRequest(spaceName).
			Key([]any{poll.ID}).
			Operations(tarantool.NewOperations().Assign(2, poll.Options)),
	).Get()
	if err != nil {
		log.Printf("[Tarantool]: Failed to update poll: %s", err.Error())
		return fmt.Errorf("[Tarantool]: Failed to update poll")
	}
	return nil
}

func Delete(config *utils.Config, poll *Poll) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address: config.DBHost + ":" + config.DBPort,
		User:    user,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		log.Println("[Tarantool]: Connection refused:", err)
		return fmt.Errorf("[Tarantool]: failed to connect to Tarantool")
	}

	_, err = conn.Do(
		tarantool.NewDeleteRequest(spaceName).
			Key([]any{poll.ID}),
	).Get()
	if err != nil {
		log.Printf("[Tarantool]: Failed to delete poll: %s", err.Error())
		return fmt.Errorf("[Tarantool]: Failed to update poll")
	}
	log.Printf("[Tarantool]: Successfully deleted poll")
	return nil
}
