package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"

	transport "sourcecraft.dev/benzo/testengine/internal/delivery/grpc"

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

	transport.RegisterSurveyAdminServiceServer(grpcServer)

	return &App{log: log, cfg: cfg, grpcServer: grpcServer}, nil
}

func (a *App) Start(ctx context.Context) error {
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

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	ch := make(chan struct{})

	go func() {
		a.grpcServer.GracefulStop()

		close(ch)
	}()

	select {
	case <-ctx.Done():
		a.grpcServer.Stop()
		return ctx.Err()
	case <-ch:
		return nil
	}
}
