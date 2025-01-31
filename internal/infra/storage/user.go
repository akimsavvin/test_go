package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/akimsavvin/efgo"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/akimsavvin/test_go/internal/usecase"
	"github.com/akimsavvin/test_go/pkg/changetracker"
	"github.com/akimsavvin/test_go/pkg/sl"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type userSnapshot struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
}

func userFromSnapshot(snap *userSnapshot) *domain.User {
	return domain.NewUser(snap.ID, snap.CreatedAt, snap.UpdatedAt, snap.Name, snap.Email)
}

type UserRepo struct {
	log  *slog.Logger
	qx   QueryExec
	coll *changetracker.EntityCollection[domain.User]
}

var _ usecase.UserRepo = (*UserRepo)(nil)

func NewUserRepo(log *slog.Logger, qx QueryExec, ct *changetracker.ChangeTracker) *UserRepo {
	repo := &UserRepo{
		log: log,
		qx:  qx,
	}

	if ct != nil {
		repo.coll = changetracker.Entity[domain.User](ct)
	}

	return repo
}

func (repo *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, created_at, updated_at, name, email FROM users WHERE id = $1;`

	snap, err := efgo.QueryRowContext[userSnapshot](ctx, repo.qx, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	user := userFromSnapshot(snap)
	if repo.coll != nil {
		repo.coll.Add(user)
	}

	return user, nil
}

func (repo *UserRepo) Insert(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, created_at, updated_at, name, email) VALUES ($1, $2, $3, $4, $5);`

	_, err := repo.qx.ExecContext(
		ctx,
		query,
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.Name(),
		user.Email())
	if err != nil {
		return err
	}

	if repo.coll != nil {
		repo.coll.Add(user)
	}

	return nil
}

func (repo *UserRepo) Remove(ctx context.Context, user *domain.User) error {
	query := `DELETE FROM users WHERE id = $1;`

	_, err := repo.qx.ExecContext(ctx, query, user.ID())
	if err != nil {
		return err
	}

	if repo.coll != nil {
		repo.coll.Remove(user)
	}

	return nil
}

func (repo *UserRepo) update(ctx context.Context, user *domain.User) error {
	log := repo.log.With(slog.String("user_id", user.ID().String()))

	query := `UPDATE users
			  SET (created_at, updated_at, name, email) = ($1, $2, $3, $4)
			  WHERE id = $5;`
	log.DebugContext(ctx, "updating in users", slog.String("query", query))

	if _, err := repo.qx.ExecContext(
		ctx,
		query,
		user.CreatedAt(),
		user.UpdatedAt(),
		user.Name(),
		user.Email(),
		user.ID(),
	); err != nil {
		log.ErrorContext(ctx, "could not update in users", sl.Err(err))
		return err
	}
	log.InfoContext(ctx, "updated in users")

	return nil
}
