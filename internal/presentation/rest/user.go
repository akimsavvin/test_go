package rest

import (
	"errors"
	"fmt"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/akimsavvin/test_go/internal/usecase"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"time"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

func userDtoToResponse(dto *usecase.UserDTO) *UserResponse {
	return &UserResponse{
		ID:        dto.ID,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		Name:      dto.Name,
		Email:     dto.Email,
	}
}

type UserController struct {
	useCase usecase.UserUseCase
}

func NewUserController(useCase usecase.UserUseCase) *UserController {
	return &UserController{
		useCase: useCase,
	}
}

func (contr *UserController) Init(root fiber.Router) {
	g := root.Group("/users")
	g.Get("/:id", contr.getById)
	g.Post("/", contr.create)
	g.Put("/:id", contr.update)
	g.Delete("/:id", contr.delete)
}

func (contr *UserController) getById(fCtx fiber.Ctx) error {
	id, err := uuid.Parse(fCtx.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	userDTO, err := contr.useCase.GetById(fCtx.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return fiber.ErrNotFound
		}

		return fiber.ErrInternalServerError
	}

	return fCtx.Status(fiber.StatusOK).JSON(userDtoToResponse(userDTO))
}

func (contr *UserController) create(fCtx fiber.Ctx) error {
	var req CreateUserRequest
	if err := fCtx.Bind().Body(&req); err != nil {
		return fiber.ErrBadRequest
	}

	dto := &usecase.CreateUserDTO{
		Name:  req.Name,
		Email: req.Email,
	}

	id, err := contr.useCase.Create(fCtx.Context(), dto)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	fCtx.Set("Content-Location", fmt.Sprintf("/api/v1/users/%s", id.String()))
	return fCtx.Status(fiber.StatusCreated).Send(nil)
}

func (contr *UserController) update(fCtx fiber.Ctx) error {
	id, err := uuid.Parse(fCtx.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	var req UpdateUserRequest
	if err = fCtx.Bind().Body(&req); err != nil {
		return fiber.ErrBadRequest
	}

	dto := &usecase.UpdateUserDTO{
		Name:  req.Name,
		Email: req.Email,
	}

	if err = contr.useCase.Update(fCtx.Context(), id, dto); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return fiber.ErrNotFound
		}

		return fiber.ErrInternalServerError
	}

	return fCtx.Status(fiber.StatusOK).Send(nil)
}

func (contr *UserController) delete(fCtx fiber.Ctx) error {
	id, err := uuid.Parse(fCtx.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err = contr.useCase.Delete(fCtx.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return fiber.ErrNotFound
		}

		return fiber.ErrInternalServerError
	}

	return fCtx.Status(fiber.StatusNoContent).Send(nil)
}
