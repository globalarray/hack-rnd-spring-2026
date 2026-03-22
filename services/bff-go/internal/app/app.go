package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"sourcecraft.dev/benzo/bff/internal/application/usecase"
	"sourcecraft.dev/benzo/bff/internal/config"
	httpapi "sourcecraft.dev/benzo/bff/internal/delivery/httpapi"
	smtpadapter "sourcecraft.dev/benzo/bff/internal/infrastructure/email/smtp"
	analyticsgrpc "sourcecraft.dev/benzo/bff/internal/infrastructure/grpc/analytics"
	enginegrpc "sourcecraft.dev/benzo/bff/internal/infrastructure/grpc/engine"
	grpcclient "sourcecraft.dev/benzo/bff/internal/infrastructure/grpc/shared"

	"google.golang.org/grpc"
)

type App struct {
	handler       http.Handler
	engineConn    *grpc.ClientConn
	analyticsConn *grpc.ClientConn
}

func New(ctx context.Context, log *slog.Logger, cfg config.Config) (*App, error) {
	engineConn, err := grpcclient.Dial(ctx, grpcclient.Config{
		Address:        cfg.Engine.Address,
		Insecure:       cfg.Engine.Insecure,
		CACertPath:     cfg.Engine.CACertPath,
		ClientCertPath: cfg.Engine.ClientCertPath,
		ClientKeyPath:  cfg.Engine.ClientKeyPath,
		ServerName:     cfg.Engine.ServerName,
	})
	if err != nil {
		return nil, fmt.Errorf("dial engine: %w", err)
	}

	analyticsConn, err := grpcclient.Dial(ctx, grpcclient.Config{
		Address:        cfg.Analytics.Address,
		Insecure:       cfg.Analytics.Insecure,
		CACertPath:     cfg.Analytics.CACertPath,
		ClientCertPath: cfg.Analytics.ClientCertPath,
		ClientKeyPath:  cfg.Analytics.ClientKeyPath,
		ServerName:     cfg.Analytics.ServerName,
	})
	if err != nil {
		_ = engineConn.Close()
		return nil, fmt.Errorf("dial analytics: %w", err)
	}

	engineGateway := enginegrpc.NewClient(engineConn)
	analyticsGateway := analyticsgrpc.NewClient(analyticsConn)
	mailer := smtpadapter.NewSender(smtpadapter.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		UseTLS:   cfg.SMTP.UseTLS,
	})

	surveyUseCase := usecase.NewSurveyUseCase(engineGateway)
	sessionUseCase := usecase.NewSessionUseCase(log, engineGateway, analyticsGateway, mailer, cfg.DefaultClientReportFormat)

	handler := httpapi.NewRouter(log, surveyUseCase, sessionUseCase)

	return &App{
		handler:       handler,
		engineConn:    engineConn,
		analyticsConn: analyticsConn,
	}, nil
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if a.analyticsConn != nil {
		if err := a.analyticsConn.Close(); err != nil {
			return err
		}
	}

	if a.engineConn != nil {
		if err := a.engineConn.Close(); err != nil {
			return err
		}
	}

	return nil
}
