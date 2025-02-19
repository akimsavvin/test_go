package changetracker

import "reflect"

type ChangeTracker struct {
	// entities are map[reflect.Type]*EntityCollection[T]
	entities map[reflect.Type]any
}

type Option func(ct *ChangeTracker)

func New(opts ...Option) *ChangeTracker {
	ct := &ChangeTracker{
		entities: make(map[reflect.Type]any),
	}

	for _, opt := range opts {
		opt(ct)
	}

	return ct
}

func WithEntity[T any](
	getKeyFunc GetEntityKeyFunc[T],
	hasChangedFunc HasEntityChangedFunc[T],
	copyFunc ...CopyEntityFunc[T]) Option {
	return func(ct *ChangeTracker) {
		typ := reflect.TypeFor[T]()
		coll := NewEntityCollection(getKeyFunc, hasChangedFunc, copyFunc...)

		ct.entities[typ] = coll
	}
}

type entity[T any] struct {
	initial *T
	current *T
}

type GetEntityKeyFunc[T any] func(*T) any

type HasEntityChangedFunc[T any] func(*T, *T) bool

type CopyEntityFunc[T any] func(*T) *T

type EntityCollection[T any] struct {
	entities       map[any]*entity[T]
	getKeyFunc     GetEntityKeyFunc[T]
	hasChangedFunc HasEntityChangedFunc[T]
	copyFunc       CopyEntityFunc[T]
}

func NewEntityCollection[T any](
	getKeyFunc GetEntityKeyFunc[T],
	hasChangedFunc HasEntityChangedFunc[T],
	copyFunc ...CopyEntityFunc[T]) *EntityCollection[T] {
	coll := &EntityCollection[T]{
		entities:       make(map[any]*entity[T]),
		getKeyFunc:     getKeyFunc,
		hasChangedFunc: hasChangedFunc,
	}

	if len(copyFunc) > 0 {
		coll.copyFunc = copyFunc[0]
	} else {
		coll.copyFunc = func(e *T) *T {
			copied := new(T)
			*copied = *e
			return copied
		}
	}

	return coll
}

func (coll *EntityCollection[T]) Add(e *T) {
	key := coll.getKeyFunc(e)
	initial := coll.copyFunc(e)
	coll.entities[key] = &entity[T]{initial, e}
}

func (coll *EntityCollection[T]) Remove(e *T) {
	key := coll.getKeyFunc(e)
	delete(coll.entities, key)
}

func (coll *EntityCollection[T]) Changed() []*T {
	res := make([]*T, 0, len(coll.entities))

	for _, e := range coll.entities {
		if coll.hasChangedFunc(e.initial, e.current) {
			res = append(res, e.current)
		}
	}

	return res
}

func Entity[T any](ct *ChangeTracker) *EntityCollection[T] {
	coll, ok := ct.entities[reflect.TypeFor[T]()]
	if !ok {
		panic("ChangeTracker: no entity collection found in change tracker")
	}

	return coll.(*EntityCollection[T])
}
