package permissions

import (
	"context"
	"log/slog"

	ssov1 "sso.service/api/proto/gen/v1"
	"sso.service/internal/entity"
)

type PermissionsService interface {
	CheckPermission(ctx context.Context, userID int64, permission string) (bool, error)
	GrantPermissions(ctx context.Context, userID int64, permissionCodes []string) ([]entity.Permission, error)
}

type PermissionsServer struct {
	ssov1.UnimplementedPermissionsServer
	service PermissionsService
	log     *slog.Logger
}

func New(service PermissionsService, log *slog.Logger) *PermissionsServer {
	return &PermissionsServer{service: service, log: log}
}
