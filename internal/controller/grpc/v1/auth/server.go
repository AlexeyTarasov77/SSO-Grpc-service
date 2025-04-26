package auth

import (
	"context"
	"log/slog"

	ssov1 "sso.service/api/proto/gen/v1"
	"sso.service/internal/entity"
	"sso.service/internal/services/dtos"
)

type AuthService interface {
	Login(ctx context.Context, username string, password string, appId int32) (*dtos.AuthTokens, error)
	Register(ctx context.Context, username string, password string, email string, appId int32) (*dtos.UserIDAndToken, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	GetOrCreateApp(ctx context.Context, app *entity.App) (*dtos.GetOrCreateAppDTO, error)
	RenewAccessToken(ctx context.Context, refreshToken string, appId int32) (string, error)
	GetUser(ctx context.Context, params dtos.GetUserOptionsDTO) (*entity.User, error)
	ActivateUser(ctx context.Context, token string, appID int32) (*entity.User, error)
	NewActivationToken(ctx context.Context, email string, appID int32) (string, error)
	VerifyToken(ctx context.Context, appID int32, token string) error
}

type AuthServer struct {
	ssov1.UnimplementedAuthServer
	service AuthService
	log     *slog.Logger
}

func New(service AuthService, log *slog.Logger) *AuthServer {
	return &AuthServer{service: service, log: log}
}
