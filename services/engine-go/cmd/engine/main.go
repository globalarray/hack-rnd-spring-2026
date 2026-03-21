package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sourcecraft.dev/benzo/testengine/internal/app"
	"sourcecraft.dev/benzo/testengine/internal/config"
	"sourcecraft.dev/benzo/testengine/pkg/log"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))

	if err != nil {
		panic(err)
	}

	l := log.New[string](cfg.Log.Level, cfg.Log.TimeFormat)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	defer cancel()

	a, err := app.New(l, cfg)

	if err != nil {
		l.With("err", err).Error("while initializing app")
		return
	}

	l.Info("setting up...")

	go func() {
		<-ctx.Done()

		l.Info("gracefully shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		if err := a.Shutdown(shutdownCtx); err != nil {
			l.With("err", err).Error("while shutting down")
		}
	}()

	if err := a.Start(ctx); err != nil {
		l.With("err", err).Error("while starting app")
	}
}
