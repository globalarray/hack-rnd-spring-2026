package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/lib/pq"
	analyticsHandler "sourcecraft.dev/benzo/testengine/internal/delivery/grpc/analytics"
	sessionHandler "sourcecraft.dev/benzo/testengine/internal/delivery/grpc/session"
	surveyHandler "sourcecraft.dev/benzo/testengine/internal/delivery/grpc/survey"
	analyticsRepo "sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/analytics"
	sessionRepo "sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/session"
	surveyRepo "sourcecraft.dev/benzo/testengine/internal/infrastructure/postgres/repository/survey"
	analyticsService "sourcecraft.dev/benzo/testengine/internal/service/analytics"
	"sourcecraft.dev/benzo/testengine/internal/service/session"

	"github.com/jmoiron/sqlx"
	"sourcecraft.dev/benzo/testengine/internal/service/survey"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"sourcecraft.dev/benzo/testengine/internal/config"
	"sourcecraft.dev/benzo/testengine/pkg/secure"
)

type App struct {
	log *slog.Logger
	cfg config.Config

	grpcServer *grpc.Server
	db         *sqlx.DB
}

func New(log *slog.Logger, cfg config.Config) (*App, error) {
	tlsCfg, err := secure.LoadTLSConfig(filepath.Join(cfg.CertsPath, "server.crt"), filepath.Join(cfg.CertsPath, "server.key"), filepath.Join(cfg.CertsPath, "ca.crt"), secure.ServerMode)

	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(recovery.UnaryServerInterceptor()),
		grpc.Creds(credentials.NewTLS(tlsCfg)),
	)

	db, err := sqlx.Open("postgres", cfg.Postgres.DSN())

	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	sRepo := surveyRepo.NewSurveyRepository(log, db)
	aRepo := analyticsRepo.NewAnalyticsRepository(log, db)

	surveyHandler.RegisterSurveyAdminServiceServer(grpcServer, log, survey.NewSurveyService(sRepo))
	sessionHandler.RegisterSessionClientServer(grpcServer, log, session.NewSessionService(log, sessionRepo.NewSessionRepository(log, db), sRepo))
	analyticsHandler.RegisterAnalyticsServiceServer(grpcServer, log, analyticsService.NewAnalyticsService(aRepo))

	return &App{log: log, cfg: cfg, grpcServer: grpcServer, db: db}, nil
}

func (a *App) Start(ctx context.Context) error {
	pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
	defer pingCancel()

	if err := a.db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("could not connect to postgres %w", err)
	}

	if err := a.startGRPCServer(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) startGRPCServer(ctx context.Context) error {
	const op = "app.startGRPCServer"

	var lc net.ListenConfig

	l, err := lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", a.cfg.Port))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.With(slog.String("addr", fmt.Sprintf("::%d", a.cfg.Port))).Info("listening grpc endpoint")

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) (err error) {
	ch := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		if dbErr := a.db.Close(); dbErr != nil {
			err = dbErr
		}
	}()

	go func() {
		defer wg.Done()

		a.grpcServer.GracefulStop()
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	select {
	case <-ctx.Done():
		a.grpcServer.Stop()
		return ctx.Err()
	case <-ch:
		return err
	}
}
