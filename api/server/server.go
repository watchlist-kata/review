package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/watchlist-kata/protos/review"
	"github.com/watchlist-kata/review/internal/config"
	"github.com/watchlist-kata/review/internal/repository"
	"github.com/watchlist-kata/review/internal/service"
)

// RunServer запускает gRPC сервер
func RunServer(ctx context.Context, cfg *config.Config) error {
	// Создание репозитория
	repo, err := repository.NewPostgresRepository(cfg)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Создание сервиса
	srv := service.NewReviewService(repo)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()

	// Регистрация сервиса
	review.RegisterReviewServiceServer(grpcServer, srv)

	// Запуск сервера
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", cfg.GRPCPort, err)
	}

	log.Printf("gRPC server listening on port %s", cfg.GRPCPort)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}
