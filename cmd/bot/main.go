package main

import (
	"context"
	"log"

	"survival-bot/internal"
	"survival-bot/internal/config"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app, err := internal.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	defer app.Close()
	if err := app.Run(ctx); err != nil {
		log.Fatalf("App error: %v", err)
	}
}
