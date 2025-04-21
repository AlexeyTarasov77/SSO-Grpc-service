package auth

import (
	"log/slog"
	"sso.service/internal/config"
)

type AuthService struct {
	log             *slog.Logger
	usersRepo       usersRepo
	appsRepo        appsRepo
	permissionsRepo permissionsRepo
	cfg             *config.Config
}

func New(
	log *slog.Logger,
	usersRepo usersRepo,
	appsRepo appsRepo,
	permissionsRepo permissionsRepo,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		log,
		usersRepo,
		appsRepo,
		permissionsRepo,
		cfg,
	}
}
