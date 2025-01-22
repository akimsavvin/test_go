package usecase

import (
	"context"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/google/uuid"
)

// UserUseCase is a use cases for domain.User
type UserUseCase interface {
	GetById(ctx context.Context, id uuid.UUID) (*UserDTO, error)
	Create(ctx context.Context, dto *CreateUserDTO) (uuid.UUID, error)
	Update(ctx context.Context, id uuid.UUID, dto *UpdateUserDTO) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepo is the domain.User repository
type UserRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Insert(ctx context.Context, user *domain.User) error
	Remove(ctx context.Context, user *domain.User) error
}

// UnitOfWork manages repository in a single unit
type UnitOfWork interface {
	Users() UserRepo

	Save() error
	Cancel() error
}

// UnitOfWorkFactory creates a new UnitOfWork
type UnitOfWorkFactory interface {
	StartWork(ctx context.Context) (UnitOfWork, error)
}
