package auth

import (
	"log/slog"
	"sso.service/internal/config"
)

type AuthService struct {
	log       *slog.Logger
	usersRepo usersRepo
	appsRepo  appsRepo
	cfg       *config.Config
}

func New(
	log *slog.Logger,
	usersRepo usersRepo,
	appsRepo appsRepo,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		log,
		usersRepo,
		appsRepo,
		cfg,
	}
}
