package app

import (
	"context"
	"log/slog"

	grpcapp "sso.service/internal/app/grpc"
	"sso.service/internal/config"
	"sso.service/internal/services/auth"
	"sso.service/internal/storage/postgres"
	"sso.service/internal/storage/postgres/models"
)


type App struct {
	GRPCApp *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config, storagePath string) *App {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DB.LoadTimeout)
	defer cancel()
	storage, err := postgres.New(ctx, storagePath)
	if err != nil {
		panic(err)
	}
	log.Info("Database connected", "path", storagePath)
	models := models.New(storage.DB)
	authService := auth.New(log, models.User, models.App, models.Permission, cfg)
	gRPCApp := grpcapp.New(log, cfg.Server.Port, authService)
	return &App{gRPCApp}
}
