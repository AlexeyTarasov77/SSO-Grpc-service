package permissions

import (
	"log/slog"
)

type PermissionsService struct {
	log             *slog.Logger
	permissionsRepo permissionsRepo
	usersRepo       usersRepo
}

func New(
	log *slog.Logger,
	permissionsRepo permissionsRepo,
	usersRepo usersRepo,
) *PermissionsService {
	return &PermissionsService{
		log,
		permissionsRepo,
		usersRepo,
	}
}
