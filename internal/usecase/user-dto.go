package usecase

import (
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/google/uuid"
	"time"
)

type UserDTO struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Email     string
}

type CreateUserDTO struct {
	Name  string
	Email string
}

type UpdateUserDTO struct {
	Name  string
	Email string
}

func userToDTO(user *domain.User) *UserDTO {
	return &UserDTO{
		ID:        user.ID(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
		Name:      user.Name(),
		Email:     user.Email(),
	}
}
