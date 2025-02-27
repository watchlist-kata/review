package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/watchlist-kata/protos/review"
	"github.com/watchlist-kata/review/internal/config"
	"github.com/watchlist-kata/review/internal/repository"
	"github.com/watchlist-kata/review/internal/service"
)

// RunServer запускает gRPC сервер
func RunServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Проверка отмены контекста
	select {
	case <-ctx.Done():
		logger.Error("server initialization canceled", slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	// Создание репозитория
	repo, err := repository.NewPostgresRepository(cfg, logger)
	if err != nil {
		logger.Error("failed to create repository", slog.Any("error", err))
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Создание сервиса
	srv := service.NewReviewService(repo, logger)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()

	// Регистрация сервиса
	review.RegisterReviewServiceServer(grpcServer, srv)

	// Запуск сервера
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen on port", slog.Any("port", cfg.GRPCPort), slog.Any("error", err))
		return fmt.Errorf("failed to listen on port %s: %w", cfg.GRPCPort, err)
	}

	logger.Info("gRPC server listening on port", slog.String("port", cfg.GRPCPort))
	fmt.Printf("gRPC server listening on port %s\n", cfg.GRPCPort)

	// Запуск сервера в отдельной горутине, чтобы не блокировать основную горутину
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve gRPC server", slog.Any("error", err))
		}
	}()

	// Ожидание завершения контекста
	<-ctx.Done()
	logger.Info("server stopped due to context cancellation")
	return ctx.Err()
}
