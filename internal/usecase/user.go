package usecase

import (
	"context"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/akimsavvin/test_go/pkg/cache"
	"github.com/akimsavvin/test_go/pkg/sl"
	"github.com/google/uuid"
	"log/slog"
)

type userUseCaseImpl struct {
	log       *slog.Logger
	ufw       UnitOfWorkFactory
	jsonCache cache.JsonCache
	evtBus    EventBus
}

func NewUserUseCase(log *slog.Logger, ufw UnitOfWorkFactory, jc cache.JsonCache, evtBus EventBus) UserUseCase {
	return &userUseCaseImpl{
		log:       log,
		ufw:       ufw,
		jsonCache: jc,
		evtBus:    evtBus,
	}
}

func (useCase *userUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	log := useCase.log.With(slog.String("user_id", id.String()))
	log.DebugContext(ctx, "getting user by id")

	dto := &UserDTO{}
	if err := useCase.jsonCache.Get(ctx, id.String(), dto); err == nil {
		log.InfoContext(ctx, "received cached user")
		return dto, nil
	}

	unit, err := useCase.ufw.StartReadWork(ctx)
	if err != nil {
		log.InfoContext(ctx, "could not get user by id", sl.Err(err))
		return nil, err
	}
	defer unit.Cancel()

	user, err := unit.Users().GetByID(ctx, id)
	if err != nil {
		log.InfoContext(ctx, "could not get user by id", sl.Err(err))
		return nil, err
	}

	if err = unit.Save(); err != nil {
		log.InfoContext(ctx, "could not get user by id", sl.Err(err))
		return nil, err
	}

	dto = userToDTO(user)
	log.ErrorContext(ctx, "caching user")
	if err = useCase.jsonCache.Set(ctx, id.String(), dto); err != nil {
		log.ErrorContext(ctx, "failed to cache user")
	}
	log.InfoContext(ctx, "got user by id")

	return dto, nil
}

func (useCase *userUseCaseImpl) Create(ctx context.Context, dto *CreateUserDTO) (uuid.UUID, error) {
	unit, err := useCase.ufw.StartWork(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer unit.Cancel()

	user := domain.CreateUser(dto.Name, dto.Email)
	if err = unit.Users().Insert(ctx, user); err != nil {
		return uuid.Nil, err
	}

	if err = unit.Save(); err != nil {
		return uuid.Nil, err
	}

	if err = useCase.evtBus.Publish(ctx, &domain.UserCreatedEvent{
		ID:        user.ID(),
		Name:      user.Name(),
		Email:     user.Email(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}); err != nil {
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

	if err = useCase.jsonCache.Del(ctx, id.String()); err != nil {
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

	if err = useCase.jsonCache.Del(ctx, id.String()); err != nil {
	}

	return nil
}
