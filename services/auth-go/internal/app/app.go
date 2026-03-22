package app

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	auth "sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/gen/go"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/config"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/secure"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/server"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/server/interceptors"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/service"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/storage/postgres"
)

type App struct {
	repo      *postgres.Storage
	logger    *slog.Logger
	cfg       config.Config
	tlsConfig *tls.Config
}

func New(repo *postgres.Storage, logger *slog.Logger, cfg config.Config) (*App, error) {
	tlsConfig, err := secure.LoadServerTLSConfig(
		filepath.Join(cfg.CertsPath, "server.crt"),
		filepath.Join(cfg.CertsPath, "server.key"),
		filepath.Join(cfg.CertsPath, "ca.crt"),
	)
	if err != nil {
		return nil, err
	}

	return &App{
		repo:      repo,
		logger:    logger,
		cfg:       cfg,
		tlsConfig: tlsConfig,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("starting auth gRPC server", slog.String("port", a.cfg.Port))

	authServer := server.New(service.NewAuthService(a.repo))
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RecoveryInterceptor(a.logger),
			interceptors.LoggingInterceptor(a.logger),
		),
		grpc.Creds(credentials.NewTLS(a.tlsConfig)),
	)

	auth.RegisterAuthServiceServer(grpcServer, authServer)

	lis, err := net.Listen("tcp", ":"+a.cfg.Port)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	return nil
}
