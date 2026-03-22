package main

import (
	"flag"
	"log"

	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/app"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/config"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/logger"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/storage/postgres"
)

func main() {
	cfgPath := flag.String("config", "", "optional config path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.Env, logger.Timeformat)

	repo, err := postgres.NewStorage(&cfg, appLogger)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}
	defer repo.DB.Close()

	application, err := app.New(repo, appLogger, cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
