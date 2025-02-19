package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type User struct {
	id        uuid.UUID
	createdAt time.Time
	updatedAt time.Time
	name      string
	email     string
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func NewUser(
	id uuid.UUID,
	createdAt time.Time,
	updatedAt time.Time,
	name string,
	email string) *User {
	return &User{
		id:        id,
		createdAt: createdAt,
		updatedAt: updatedAt,
		name:      name,
		email:     email,
	}
}

func CreateUser(name string, email string) *User {
	now := time.Now()
	return NewUser(uuid.New(), now, now, name, email)
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Email() string {
	return u.email
}

func (u *User) Update(name, email string) {
	u.name = name
	u.email = email
	u.updatedAt = time.Now()
}

type UserCreatedEvent struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}
