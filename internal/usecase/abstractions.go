package usecase

import (
	"context"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/google/uuid"
)

// UserUseCase is a use cases for domain.User
type UserUseCase interface {
	// GetByID returns a user by identifier
	GetByID(ctx context.Context, id uuid.UUID) (*UserDTO, error)

	// Create creates a new user and returns its identifier
	Create(ctx context.Context, dto *CreateUserDTO) (uuid.UUID, error)

	// Update updates an existing user by its identifier
	Update(ctx context.Context, id uuid.UUID, dto *UpdateUserDTO) error

	// Delete deletes the user by its identifier
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserReadRepo is the domain.User read repository
type UserReadRepo interface {
	// GetByID returns a user by identifier
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// UserRepo is the domain.User repository
type UserRepo interface {
	UserReadRepo

	// Insert inserts a user into the repository
	Insert(ctx context.Context, user *domain.User) error

	// Remove removes the user from the repository
	Remove(ctx context.Context, user *domain.User) error
}

// UnitOfWorkBase contains Save and Cancel methods
type UnitOfWorkBase interface {
	// Save saves changes in the repositories
	Save() error

	// Cancel cancels the work
	Cancel() error
}

// UnitOfWork manages repositories in a single unit
type UnitOfWork interface {
	UnitOfWorkBase

	// Users returns the user repository
	Users() UserRepo
}

// UnitOfReadWork manages read repositories in a single read unit
type UnitOfReadWork interface {
	UnitOfWorkBase

	// Users returns the user read repository
	Users() UserReadRepo
}

// UnitOfWorkFactory creates a new UnitOfWork
type UnitOfWorkFactory interface {
	// StartWork starts a new unit of work
	StartWork(ctx context.Context) (UnitOfWork, error)

	// StartReadWork starts a new unit of read work
	StartReadWork(ctx context.Context) (UnitOfReadWork, error)
}
