package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/shared/domain"
)

type RedisEventBus struct {
	client *redis.Client
}

func NewRedisEventBus(client *redis.Client) *RedisEventBus {
	return &RedisEventBus{
		client: client,
	}
}

func (b *RedisEventBus) Publish(ctx context.Context, event domain.DomainEvent) error {
	channel := fmt.Sprintf("events:%s", event.Name())

	// We wrap the event to include metadata if needed, but for now just marshaling the event itself
	// Event structures must be JSON serializable.
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return b.client.Publish(ctx, channel, data).Err()
}

func (b *RedisEventBus) Subscribe(ctx context.Context, eventName string, handler func([]byte) error) {
	channel := fmt.Sprintf("events:%s", eventName)
	pubsub := b.client.Subscribe(ctx, channel)

	go func() {
		defer func() { _ = pubsub.Close() }()
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if msg == nil {
					continue
				}
				if err := handler([]byte(msg.Payload)); err != nil {
					log.Printf("Error handling event %s: %v", eventName, err)
				}
			}
		}
	}()
}
