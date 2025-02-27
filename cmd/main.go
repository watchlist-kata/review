package main

import (
	"context"
	"github.com/watchlist-kata/review/api/server"
	"github.com/watchlist-kata/review/internal/config"
	"github.com/watchlist-kata/review/pkg/logger"
	"log"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация кастомного логгера
	customLogger, err := logger.NewLogger(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.ServiceName, cfg.LogBufferSize)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if multiHandler, ok := customLogger.Handler().(*logger.MultiHandler); ok {
			multiHandler.CloseAll()
		}
	}()

	// Запуск сервера
	if err = server.RunServer(context.Background(), cfg, customLogger); err != nil {
		log.Fatal(err)
	}
}
