package app

import (
	"context"
	"database/sql"
	"errors"
	"github.com/akimsavvin/gonet/v2/di"
	"github.com/akimsavvin/test_go/internal/infra/config"
	"github.com/akimsavvin/test_go/internal/infra/storage"
	"github.com/akimsavvin/test_go/internal/presentation/rest"
	"github.com/akimsavvin/test_go/internal/usecase"
	"github.com/akimsavvin/test_go/pkg/sl"
	"github.com/gofiber/fiber/v3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	"os"
)

type App struct {
	ctx context.Context
	err error
	log *slog.Logger
	cfg config.Config
}

func New(ctx context.Context) *App {
	return &App{
		ctx: ctx,
	}
}

func (app *App) Configure() *App {
	app.log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	if err := cleanenv.ReadConfig("./config/config.yaml", &app.cfg); err != nil {
		app.err = err
		return app
	}

	return app
}

func (app *App) AddServices() *App {
	if app.err != nil {
		return app
	}

	di.AddValue(app.log)

	di.AddKeyedFactory("master", func() (*sql.DB, error) {
		return sql.Open("pgx", app.cfg.DB.MasterURL)
	})
	di.AddKeyedFactory("slave", func() (*sql.DB, error) {
		return sql.Open("pgx", app.cfg.DB.SlaveURL)
	})

	di.AddFactory(func(log *slog.Logger, sp di.ServiceProvider) (usecase.UnitOfWorkFactory, error) {
		master := di.GetRequiredKeyedServiceSP[*sql.DB](sp, "master")
		slave := di.GetRequiredKeyedServiceSP[*sql.DB](sp, "slave")
		return storage.NewUnitOfWorkFactory(log, master, slave), nil
	})

	di.AddFactory(usecase.NewUserUseCase)
	di.AddService[rest.Controller](rest.NewUserController)

	return app
}

func (app *App) Run() error {
	if app.err != nil {
		return app.err
	}

	di.Build()

	log := app.log.With(sl.Op("app.Run"))

	db := di.GetRequiredService[*sql.DB]()
	log.Debug("migrating database")
	if err := storage.Migrate(db); err != nil {
		log.Debug("could not migrate database", sl.Err(err))
		return err
	}
	log.Info("migrated database")

	g, _ := errgroup.WithContext(app.ctx)

	g.Go(func() error {
		log := log.With(slog.String("address", app.cfg.RestServer.Addr))
		log.Debug("starting REST server")

		fiberApp := fiber.New()

		root := fiberApp.Group("/api/v1")
		for _, c := range di.GetRequiredService[[]rest.Controller]() {
			c.Init(root)
		}

		if err := fiberApp.Listen(app.cfg.RestServer.Addr); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("REST server stopped")
			} else {
				log.Error("could not start REST server", err.Error())
			}

			return err
		}

		return nil
	})

	return g.Wait()
}
