package kfk

import (
	"context"
	"errors"
)

var (
	ErrConsumerStopped = errors.New("consumer has been stopped")
)

type Consumer interface {
	Run(ctx context.Context) error
}

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}
