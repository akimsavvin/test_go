package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/akimsavvin/test_go/internal/usecase"
	"github.com/akimsavvin/test_go/pkg/changetracker"
	"github.com/akimsavvin/test_go/pkg/sl"
	"log/slog"
)

type UnitOfWork struct {
	ctx context.Context

	log *slog.Logger
	tx  *sql.Tx

	userRepo *UserRepo

	ct *changetracker.ChangeTracker
}

var _ usecase.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(ctx context.Context, log *slog.Logger, tx *sql.Tx) *UnitOfWork {
	ct := changetracker.New(
		changetracker.WithEntity[domain.User](
			func(user *domain.User) any {
				return user.ID()
			},
			func(old *domain.User, current *domain.User) bool {
				return old.UpdatedAt() != current.UpdatedAt()
			},
		),
	)

	return &UnitOfWork{
		ctx: ctx,
		log: log,
		tx:  tx,
		ct:  ct,
	}
}

func (unit *UnitOfWork) Users() usecase.UserRepo {
	if unit.userRepo == nil {
		unit.userRepo = NewUserRepo(unit.log, unit.tx, unit.ct)
	}

	return unit.userRepo
}

func (unit *UnitOfWork) Save() (err error) {
	log := unit.log.With(sl.Op("Save"))

	usersColl := changetracker.Entity[domain.User](unit.ct)
	for _, user := range usersColl.Changed() {
		if err = unit.userRepo.update(unit.ctx, user); err != nil {
			return err
		}
	}

	if err = unit.tx.Commit(); err != nil {
		if errors.Is(err, sql.ErrTxDone) {
			log.Debug("work is already finished")
			return nil
		} else {
			log.Error("could not commit unit of work", sl.Err(err))
			return err
		}
	}

	log.Info("commited unit of work")
	return nil
}

func (unit *UnitOfWork) Cancel() error {
	log := unit.log.With(sl.Op("Cancel"))

	if err := unit.tx.Rollback(); err != nil {
		if errors.Is(err, sql.ErrTxDone) {
			log.Debug("work is already finished")
			return nil
		} else {
			log.Error("could not rollback unit of work", sl.Err(err))
			return err
		}
	}

	log.Info("rolled unit of work back")
	return nil
}

type UnitOfWorkFactory struct {
	log *slog.Logger
	db  *sql.DB
}

var _ usecase.UnitOfWorkFactory = (*UnitOfWorkFactory)(nil)

func NewUnitOfWorkFactory(log *slog.Logger, db *sql.DB) *UnitOfWorkFactory {
	return &UnitOfWorkFactory{
		log: log,
		db:  db,
	}
}

func (factory *UnitOfWorkFactory) StartWork(ctx context.Context) (usecase.UnitOfWork, error) {
	log := factory.log.With(sl.Op("StartWork"))
	log.DebugContext(ctx, "starting new unit of work")

	tx, err := factory.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		log.ErrorContext(ctx, "could not start new unit of work", sl.Err(err))
		return nil, err
	}

	log.InfoContext(ctx, "started new unit of work")
	return NewUnitOfWork(ctx, factory.log, tx), nil
}
