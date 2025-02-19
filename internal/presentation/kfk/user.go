package kfk

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/akimsavvin/test_go/internal/usecase"
	"github.com/akimsavvin/test_go/pkg/sl"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"strings"
)

type CreateMessagePayload struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserConsumer struct {
	log     *slog.Logger
	cfg     ConsumerConfig
	useCase usecase.UserUseCase
}

var _ Consumer = (*CreateUserConsumer)(nil)

func NewCreateUserConsumer(
	log *slog.Logger,
	cfg ConsumerConfig,
	useCase usecase.UserUseCase) *CreateUserConsumer {
	log = log.With(slog.Group(
		"consumer",
		slog.String("brokers", strings.Join(cfg.Brokers, ",")),
		slog.String("group_id", cfg.GroupID),
		slog.String("topic", cfg.Topic),
	))

	return &CreateUserConsumer{
		log:     log,
		cfg:     cfg,
		useCase: useCase,
	}
}

func (cons *CreateUserConsumer) Run(ctx context.Context) error {
	cons.log.InfoContext(ctx, "running consumer")

	read := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cons.cfg.Brokers,
		GroupID:  cons.cfg.GroupID,
		Topic:    cons.cfg.Topic,
		MaxBytes: 10e6, // 10MB
	})
	defer read.Close()

	for {
		msg, err := read.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				cons.log.InfoContext(ctx, "consumer stopped", sl.Err(err))
				return ErrConsumerStopped
			}

			cons.log.ErrorContext(ctx, "failed to read message", sl.Err(err))
			return err
		}

		var payload CreateMessagePayload
		if err = json.Unmarshal(msg.Value, &payload); err != nil {
			cons.log.ErrorContext(ctx, "failed to unmarshal message value", sl.Err(err))
			continue
		}

		_, _ = cons.useCase.Create(ctx, &usecase.CreateUserDTO{
			Name:  payload.Name,
			Email: payload.Email,
		})
	}
}
