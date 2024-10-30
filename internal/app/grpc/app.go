package grpcapp

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	authgrpc "sso.service/internal/grpc/auth"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpc.Server
	healthChecker *health.Server
	Host       string
	Port       string
}

const system = "" // used in health check to indicate state of whole system

func New(log *slog.Logger, host string, port string, auth authgrpc.AuthService) *App {
	gRPCServer := grpc.NewServer()
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(gRPCServer, healthcheck)
	authgrpc.Register(gRPCServer, auth, log)
	return &App{log, gRPCServer, healthcheck, host, port}
}

func (app *App) Run() {
	serverAddr := net.JoinHostPort(app.Host, app.Port)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		app.log.Error("Failed to listen", "error", err, "port", app.Port)
		panic(err)
	}
	app.log.Info("Starting gRPC server", "listener", listener.Addr(), "address", serverAddr)
	app.healthChecker.SetServingStatus(system, healthgrpc.HealthCheckResponse_SERVING)
	if err := app.GRPCServer.Serve(listener); err != nil {
		app.log.Error("Failed to serve gRPC server", "error", err)
		app.healthChecker.SetServingStatus(system, healthgrpc.HealthCheckResponse_NOT_SERVING)
		panic(err)
	}
}

func (app *App) Stop() {
	app.log.Info("Stopping gRPC server")
	app.healthChecker.SetServingStatus(system, healthgrpc.HealthCheckResponse_NOT_SERVING)
	app.GRPCServer.GracefulStop()
}
