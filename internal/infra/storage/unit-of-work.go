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
		changetracker.WithEntity(
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

func (unit *UnitOfWork) Save() error {
	log := unit.log.With(sl.Op("Save"))
	log.Debug("saving unit of work")

	var err error
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
			log.Error("could not save unit of work", sl.Err(err))
			return err
		}
	}

	log.Info("saved unit of work")
	return nil
}

func (unit *UnitOfWork) Cancel() error {
	log := unit.log.With(sl.Op("Cancel"))
	log.Debug("cancelling unit of work")

	if err := unit.tx.Rollback(); err != nil {
		if errors.Is(err, sql.ErrTxDone) {
			log.Debug("work is already finished")
			return nil
		} else {
			log.Error("could not cancel unit of work", sl.Err(err))
			return err
		}
	}

	log.Info("cancelled unit of work back")
	return nil
}

type UnitOfReadWork struct {
	ctx context.Context

	log *slog.Logger
	tx  *sql.Tx

	userRepo *UserRepo
}

var _ usecase.UnitOfReadWork = (*UnitOfReadWork)(nil)

func NewUnitOfReadWork(ctx context.Context, log *slog.Logger, tx *sql.Tx) *UnitOfReadWork {
	return &UnitOfReadWork{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

func (unit *UnitOfReadWork) Users() usecase.UserReadRepo {
	if unit.userRepo == nil {
		unit.userRepo = NewUserRepo(unit.log, unit.tx, nil)
	}

	return unit.userRepo
}

func (unit *UnitOfReadWork) Save() error {
	log := unit.log.With(sl.Op("Save"))

	if err := unit.tx.Commit(); err != nil {
		if errors.Is(err, sql.ErrTxDone) {
			log.Debug("read work is already finished")
			return nil
		} else {
			log.Error("could not save unit of read work", sl.Err(err))
			return err
		}
	}

	log.Info("saved unit of read work")
	return nil
}

func (unit *UnitOfReadWork) Cancel() error {
	log := unit.log.With(sl.Op("Cancel"))

	if err := unit.tx.Rollback(); err != nil {
		if errors.Is(err, sql.ErrTxDone) {
			log.Debug("read work is already finished")
			return nil
		} else {
			log.Error("could not cancel unit of read work", sl.Err(err))
			return err
		}
	}

	log.Info("canceled unit of read work")
	return nil
}

type UnitOfWorkFactory struct {
	log    *slog.Logger
	master *sql.DB
	slave  *sql.DB
}

var _ usecase.UnitOfWorkFactory = (*UnitOfWorkFactory)(nil)

func NewUnitOfWorkFactory(log *slog.Logger, master, slave *sql.DB) *UnitOfWorkFactory {
	return &UnitOfWorkFactory{
		log:    log,
		master: master,
		slave:  slave,
	}
}

func (factory *UnitOfWorkFactory) StartWork(ctx context.Context) (usecase.UnitOfWork, error) {
	log := factory.log.With(sl.Op("StartWork"))
	log.DebugContext(ctx, "starting new unit of work")

	tx, err := factory.master.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		log.ErrorContext(ctx, "could not start new unit of work", sl.Err(err))
		return nil, err
	}

	log.InfoContext(ctx, "started new unit of work")
	return NewUnitOfWork(ctx, factory.log, tx), nil
}

func (factory *UnitOfWorkFactory) StartReadWork(ctx context.Context) (usecase.UnitOfReadWork, error) {
	log := factory.log.With(sl.Op("StartReadWork"))
	log.DebugContext(ctx, "starting new unit of read work")

	tx, err := factory.slave.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  true,
	})
	if err != nil {
		log.ErrorContext(ctx, "could not start new unit of read work", sl.Err(err))
		return nil, err
	}

	log.InfoContext(ctx, "started new unit of read work")
	return NewUnitOfReadWork(ctx, log, tx), nil
}
