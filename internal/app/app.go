package app

import (
	"log/slog"

	grpcapp "sso.service/internal/app/grpc"
	"sso.service/internal/config"
	"sso.service/internal/services/auth"
	"sso.service/internal/storage/postgres"
)


type App struct {
	GRPCApp *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config, storagePath string) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}
	authService := auth.New(log, storage, storage, storage, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	gRPCApp := grpcapp.New(log, cfg.Server.Port, authService)
	return &App{gRPCApp}
}
