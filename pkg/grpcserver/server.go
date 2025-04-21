package grpcserver

import (
	"log/slog"
	"net"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	log           *slog.Logger
	GRPCServer    *grpc.Server
	ServeErr      chan error
	healthChecker *health.Server
	Host          string
	Port          string
}

const fullSystemHealthServing = ""

func New(log *slog.Logger, host string, port string, authServer ssov1.AuthServer) *Server {
	gRPCServer := grpc.NewServer()
	ssov1.RegisterAuthServer(gRPCServer, authServer)
	healthcheckServer := health.NewServer()
	healthgrpc.RegisterHealthServer(gRPCServer, healthcheckServer)
	return &Server{log, gRPCServer, make(chan error, 1), healthcheckServer, host, port}
}

func (self *Server) Run() {
	serverAddr := net.JoinHostPort(self.Host, self.Port)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		self.log.Error("Failed to listen", "error", err, "port", self.Port)
		panic(err)
	}
	self.log.Info("Starting gRPC server", "listener", listener.Addr(), "address", serverAddr)
	self.healthChecker.SetServingStatus(fullSystemHealthServing, healthgrpc.HealthCheckResponse_SERVING)
	if err := self.GRPCServer.Serve(listener); err != nil {
		self.log.Error("Failed to serve gRPC server", "error", err)
		self.healthChecker.SetServingStatus(fullSystemHealthServing, healthgrpc.HealthCheckResponse_NOT_SERVING)
		self.ServeErr <- err
	}
}

func (app *Server) Stop() {
	app.log.Info("Stopping gRPC server")
	app.healthChecker.SetServingStatus(fullSystemHealthServing, healthgrpc.HealthCheckResponse_NOT_SERVING)
	app.GRPCServer.GracefulStop()
}
