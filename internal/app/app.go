package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sso.service/internal/config"
	grpcV1 "sso.service/internal/controller/grpc/v1"
	"sso.service/internal/services/auth"
	"sso.service/internal/services/permissions"
	"sso.service/internal/storage/postgres"
	"sso.service/internal/storage/postgres/models"
	"sso.service/pkg/grpcserver"
)

func Run(log *slog.Logger, cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	storage, err := postgres.New(ctx, cfg.DB.Dsn)
	if err != nil {
		panic(err)
	}
	log.Info("Database connected", "dsn", cfg.DB.Dsn)
	models := models.New(storage.DB)
	authService := auth.New(log, models.User, models.App, cfg)
	permissionsService := permissions.New(log, models.Permission, models.User)
	servers := grpcV1.New(authService, permissionsService, log)
	gRPCServer := grpcserver.New(log, cfg.Server.Host, cfg.Server.Port, servers.AuthServer, servers.PermissionsServer)
	go gRPCServer.Run()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	select {
	case s := <-stop:
		log.Info("Received signal", "name", s.String())
	case <-gRPCServer.ServeErr:
		log.Info("gRPC serve error")
	}
	log.Info("Shutting down...")
	gRPCServer.Stop()
}
