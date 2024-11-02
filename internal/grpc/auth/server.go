package auth

import (
	"context"
	"log/slog"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc"
	"sso.service/internal/domain/models"
	"sso.service/internal/services/auth"
)

type AuthService interface {
	Login(ctx context.Context, username string, password string, appId int32) (*models.AuthTokens, error)
	Register(ctx context.Context, username string, password string, email string, appId int32) (*auth.UserIDAndToken, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	GetOrCreateApp(ctx context.Context, app *models.App) (*auth.AppIDAndIsCreated, error)
	RenewAccessToken(ctx context.Context, refreshToken string, appId int32) (string, error)
	GetUser(ctx context.Context, params auth.GetUserParams) (*models.User, error)
	ActivateUser(ctx context.Context, token string, appID int32) (*models.User, error)
	NewActivationToken(ctx context.Context, email string, appID int32) (string, error)
	VerifyToken(ctx context.Context, appID int32, token string) error
	CreatePermission(ctx context.Context, code string) (*models.Permission, error)
	CheckPermission(ctx context.Context, userID int64, permission string) (bool, error)
	GrantPermissions(ctx context.Context, userID int64, permissionCodes ...string) error
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth AuthService
	log  *slog.Logger
}

func Register(gRPC *grpc.Server, auth AuthService, log *slog.Logger) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth, log: log})
}

// TODO: made a health check endpoint
