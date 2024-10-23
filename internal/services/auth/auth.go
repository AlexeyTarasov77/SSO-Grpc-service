package auth

import (
	"log/slog"
	"sso.service/internal/config"
)


type Auth struct {
	log          *slog.Logger
	usersModel usersModel
	appsModel  appsModel
	permissionsModel permissionsModel
	cfg          *config.Config
}

func New(
	log *slog.Logger,
	userModel usersModel,
	appsModel appsModel,
	permissionsModel permissionsModel,
	cfg *config.Config,
) *Auth {
	return &Auth{
		log,
		userModel,
		appsModel,
		permissionsModel,
		cfg,
	}
}
