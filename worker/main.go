package main

import (
	"context"
	"log"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/worker/infrastructure"
	"github.com/rrbarrero/justbackup/internal/worker/monitoring"
)

func main() {
	log.Println("Starting Backup Worker...")

	// Load and validate configuration
	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	queueName := "backup_tasks"
	resultQueue := "backup_results"

	// Start stats collector in background
	collector := monitoring.NewStatsCollector(cfg)
	go collector.Start(ctx)

	consumer := infrastructure.NewRedisTaskConsumer(cfg.RedisURL, queueName, resultQueue)
	consumer.Start(ctx)
}
