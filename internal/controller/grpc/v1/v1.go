package v1

import (
	"context"
	"log/slog"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"sso.service/internal/entity"
	"sso.service/internal/services/auth"
)

type authService interface {
	Login(ctx context.Context, username string, password string, appId int32) (*entity.AuthTokens, error)
	Register(ctx context.Context, username string, password string, email string, appId int32) (*auth.UserIDAndToken, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	GetOrCreateApp(ctx context.Context, app *entity.App) (*auth.GetOrCreateAppDTO, error)
	RenewAccessToken(ctx context.Context, refreshToken string, appId int32) (string, error)
	GetUser(ctx context.Context, params auth.GetUserOptionsDTO) (*entity.User, error)
	ActivateUser(ctx context.Context, token string, appID int32) (*entity.User, error)
	NewActivationToken(ctx context.Context, email string, appID int32) (string, error)
	VerifyToken(ctx context.Context, appID int32, token string) error
	CreatePermission(ctx context.Context, code string) (*entity.Permission, error)
	CheckPermission(ctx context.Context, userID int64, permission string) (bool, error)
	GrantPermissions(ctx context.Context, userID int64, permissionCodes ...string) error
}

type authServer struct {
	ssov1.UnimplementedAuthServer
	service authService
	log     *slog.Logger
}

func New(service authService, log *slog.Logger) *authServer {
	return &authServer{service: service, log: log}
}
