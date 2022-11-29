package main

import (
	"context"

	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/v9"
)

var redis_client *redis.Client

func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func main() {
	ctx := context.Background()
	subscriber := redis_client.Subscribe(ctx, "doggos")
	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			logger.Errorw("Failed to receive message", "error", err)
		}

		// TODO uid will be inside payload, unmarsal this first, same for expiry time
		uid_string := msg.Payload
		err = redis_client.Set(context.Background(), uid_string, "processing", 3600).Err()
		if err != nil {
			logger.Errorw("Failed to update request status", "error", err)
			continue
		}

		// TODO do not give msg, domain logic should be independent of implementation details
		process(msg)

		err = redis_client.Set(context.Background(), uid_string, "done", 3600).Err()
		if err != nil {
			logger.Errorw("Failed to update request status", "error", err)
			continue
		}
	}
}

func process(_ *redis.Message) {
}
