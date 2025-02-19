package eventbus

import (
	"context"
	"errors"
	"github.com/akimsavvin/test_go/internal/usecase"
	"log/slog"
	"reflect"
)

var (
	ErrEventNotRegistered = errors.New("event is not registered")
)

type Publisher[TEvent any] interface {
	Publish(ctx context.Context, event TEvent) error
}

type Option func(*EventBus)

func WithEventPublisher[TEvent any](pub Publisher[TEvent]) Option {
	return func(bus *EventBus) {
		bus.pubs[reflect.TypeFor[TEvent]()] = reflect.ValueOf(pub.Publish)
	}
}

type EventBus struct {
	log  *slog.Logger
	pubs map[reflect.Type]reflect.Value
}

var _ usecase.EventBus = (*EventBus)(nil)

func New(log *slog.Logger, opts ...Option) *EventBus {
	bus := &EventBus{
		log:  log,
		pubs: make(map[reflect.Type]reflect.Value),
	}

	for _, opt := range opts {
		opt(bus)
	}

	return bus
}

func (bus *EventBus) Publish(ctx context.Context, event any) error {
	pubFunc, ok := bus.pubs[reflect.TypeOf(event)]
	if !ok {
		return ErrEventNotRegistered
	}

	return pubFunc.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(event),
	})[0].Interface().(error)
}
