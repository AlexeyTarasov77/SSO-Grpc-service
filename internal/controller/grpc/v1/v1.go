package v1

import (
	"log/slog"

	"sso.service/internal/controller/grpc/v1/auth"
	"sso.service/internal/controller/grpc/v1/permissions"
)

type GRPCServers struct {
	AuthServer        *auth.AuthServer
	PermissionsServer *permissions.PermissionsServer
}

func New(authService auth.AuthService, permissionsService permissions.PermissionsService, log *slog.Logger) *GRPCServers {
	return &GRPCServers{
		AuthServer:        auth.New(authService, log),
		PermissionsServer: permissions.New(permissionsService, log),
	}
}
