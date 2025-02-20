package app

import (
	"context"
	"database/sql"
	"errors"
	"github.com/akimsavvin/gonet/v2/di"
	"github.com/akimsavvin/test_go/internal/domain"
	"github.com/akimsavvin/test_go/internal/infra/config"
	"github.com/akimsavvin/test_go/internal/infra/eventbus"
	"github.com/akimsavvin/test_go/internal/infra/storage"
	"github.com/akimsavvin/test_go/internal/presentation/kfk"
	"github.com/akimsavvin/test_go/internal/presentation/rest"
	"github.com/akimsavvin/test_go/internal/usecase"
	"github.com/akimsavvin/test_go/pkg/cache"
	"github.com/akimsavvin/test_go/pkg/sl"
	"github.com/gofiber/fiber/v3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	"os"
)

func Run(ctx context.Context) error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	var cfg config.Config
	if err := cleanenv.ReadConfig("./config/config.yaml", &cfg); err != nil {
		return err
	}

	c := di.NewContainer(
		di.WithValue(log),
		di.WithKeyedFactory("master", func() (*sql.DB, error) {
			return sql.Open("pgx", cfg.DB.MasterURL)
		}),
		di.WithKeyedFactory("slave", func() (*sql.DB, error) {
			return sql.Open("pgx", cfg.DB.SlaveURL)
		}),
		di.WithFactory(func(c *di.Container) (*sql.DB, error) {
			return di.GetKeyedService[*sql.DB](c, "master")
		}),
		di.WithFactory(func(log *slog.Logger, c *di.Container) (usecase.UnitOfWorkFactory, error) {
			master := di.MustGetKeyedService[*sql.DB](c, "master")
			slave := di.MustGetKeyedService[*sql.DB](c, "slave")
			return storage.NewUnitOfWorkFactory(log, master, slave), nil
		}),
		di.WithValue(redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})),
		di.WithService[cache.JsonCache](cache.NewRedisJsonCache),
		di.WithFactory(func(log *slog.Logger) eventbus.Publisher[*domain.UserCreatedEvent] {
			return eventbus.NewUserCreatedEventPublisher(log, &kafka.Writer{
				Addr:  kafka.TCP(cfg.UserCreatedPub.Brokers...),
				Topic: cfg.UserCreatedPub.Topic,
			})
		}),
		di.WithFactory(func(
			log *slog.Logger,
			userCreatedPub eventbus.Publisher[*domain.UserCreatedEvent],
		) usecase.EventBus {
			return eventbus.New(log,
				eventbus.WithEventPublisher(userCreatedPub),
			)
		}),
		di.WithFactory(usecase.NewUserUseCase),
		di.WithService[rest.Controller](rest.NewUserController),
		di.WithService[kfk.Consumer](func(log *slog.Logger, useCase usecase.UserUseCase) *kfk.CreateUserConsumer {
			consCfg := kfk.ConsumerConfig{
				Brokers: cfg.CreateUserCons.Brokers,
				Topic:   cfg.CreateUserCons.Topic,
				GroupID: cfg.CreateUserCons.Topic,
			}

			return kfk.NewCreateUserConsumer(log, consCfg, useCase)
		}),
	)

	log = log.With(sl.Op("app.Run"))

	db := di.MustGetService[*sql.DB](c)
	log.Debug("migrating database")
	if err := storage.Migrate(db); err != nil {
		log.Debug("could not migrate database", sl.Err(err))
		return err
	}
	log.Info("migrated database")

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log := log.With(slog.String("address", cfg.RestServer.Addr))
		log.Debug("starting REST server")

		fiberApp := fiber.New()
		v1 := fiberApp.Group("/api/v1")

		for _, cont := range di.MustGetService[[]rest.Controller](c) {
			cont.Init(v1)
		}

		if err := fiberApp.Listen(cfg.RestServer.Addr); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("REST server stopped")
			} else {
				log.Error("could not start REST server", err.Error())
			}

			return err
		}

		return nil
	})

	log.Debug("starting kafka consumers")
	for _, cons := range di.MustGetService[[]kfk.Consumer](c) {
		g.Go(func() error {
			return cons.Run(ctx)
		})
	}

	return g.Wait()
}
