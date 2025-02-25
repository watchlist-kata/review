package main

import (
	"context"
	"log"

	"github.com/watchlist-kata/review/api/server"
	"github.com/watchlist-kata/review/internal/config"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Запуск сервера
	if err := server.RunServer(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
}
