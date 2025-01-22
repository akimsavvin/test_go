package usecase

import (
	"context"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/google/uuid"
	"log/slog"
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

type userUseCaseImpl struct {
	log *slog.Logger
	ufw UnitOfWorkFactory
}

func NewUserUseCase(log *slog.Logger, ufw UnitOfWorkFactory) UserUseCase {
	return &userUseCaseImpl{
		log: log,
		ufw: ufw,
	}
}

func (useCase *userUseCaseImpl) GetById(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	unit, err := useCase.ufw.StartWork(ctx)
	if err != nil {
		return nil, err
	}
	defer unit.Cancel()

	user, err := unit.Users().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = unit.Save(); err != nil {
		return nil, err
	}

	return userToDTO(user), nil
}

func (useCase *userUseCaseImpl) Create(ctx context.Context, dto *CreateUserDTO) (uuid.UUID, error) {
	unit, err := useCase.ufw.StartWork(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer unit.Cancel()

	user := domain.CreateUser(dto.Name, dto.Email)
	if err := unit.Users().Insert(ctx, user); err != nil {
		return uuid.Nil, err
	}

	if err = unit.Save(); err != nil {
		return uuid.Nil, err
	}

	return user.ID(), nil
}

func (useCase *userUseCaseImpl) Update(ctx context.Context, id uuid.UUID, dto *UpdateUserDTO) error {
	unit, err := useCase.ufw.StartWork(ctx)
	if err != nil {
		return err
	}
	defer unit.Cancel()

	user, err := unit.Users().GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Update(dto.Name, dto.Email)

	if err = unit.Save(); err != nil {
		return err
	}

	return nil
}

func (useCase *userUseCaseImpl) Delete(ctx context.Context, id uuid.UUID) error {
	unit, err := useCase.ufw.StartWork(ctx)
	if err != nil {
		return err
	}
	defer unit.Cancel()

	user, err := unit.Users().GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err = unit.Users().Remove(ctx, user); err != nil {
		return err
	}

	if err = unit.Save(); err != nil {
		return err
	}

	return nil
}
