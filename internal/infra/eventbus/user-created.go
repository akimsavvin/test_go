package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/segmentio/kafka-go"
	"log/slog"
)

type UserCreatedEventPublisher struct {
	log *slog.Logger
	w   *kafka.Writer
}

var _ Publisher[*domain.UserCreatedEvent] = (*UserCreatedEventPublisher)(nil)

func NewUserCreatedEventPublisher(log *slog.Logger, w *kafka.Writer) *UserCreatedEventPublisher {
	return &UserCreatedEventPublisher{
		log: log,
		w:   w,
	}
}

func (pub *UserCreatedEventPublisher) Publish(ctx context.Context, event *domain.UserCreatedEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling event: %w", err)
	}

	key, _ := event.ID.MarshalText()
	return pub.w.WriteMessages(context.Background(), kafka.Message{
		Key:   key,
		Value: bytes,
	})
}
