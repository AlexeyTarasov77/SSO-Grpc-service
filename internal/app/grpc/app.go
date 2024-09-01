package grpcapp

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	authgrpc "sso.service/internal/grpc/auth"
)

type App struct {
	log *slog.Logger
	GRPCServer *grpc.Server
	Port string
}

func New(log *slog.Logger, port string, auth authgrpc.Auth) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.Register(gRPCServer, auth)
	return &App{log, gRPCServer, port}
}

func (app *App) Run() {
	listener, err := net.Listen("tcp", ":" + app.Port)
	if err != nil {
		app.log.Error("Failed to listen", "error", err, "port", app.Port)
		panic(err)
	}
	app.log.Info("Starting gRPC server", "listener", listener.Addr())
	if err := app.GRPCServer.Serve(listener); err != nil {
		app.log.Error("Failed to serve gRPC server", "error", err)
		panic(err)
	}
}

func (app *App) Stop() {
	app.log.Info("Stopping gRPC server")
	app.GRPCServer.GracefulStop()
}