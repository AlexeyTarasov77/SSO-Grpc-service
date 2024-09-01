package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/domain/models"
	"sso.service/internal/lib/validator"
	"sso.service/internal/services/auth"
)

type Auth interface {
	Login(ctx context.Context, username string, password string, appId int32) (models.Tokens, error)
	Register(ctx context.Context, username string, password string, email string) (int64, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	GetOrCreateApp(ctx context.Context,app *models.App) (int64, bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
	log  *slog.Logger
}

func Register(gRPC *grpc.Server, auth Auth, log *slog.Logger) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth, log: log})
}
func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	validationRules := map[string]string{
		"Email": "required,email",
		"Password": "required,min=8",
		"AppId":   "required,gt=0",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		s.log.Debug("Validation errors at login", "errors", errs)
		return nil, status.Error(codes.InvalidArgument, errs)
	}

	tokens, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &ssov1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to check if user is admin")
	}
	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	validationRules := map[string]string{
		"Username": "required",
		"Password": "required,min=8",
		"Email":    "required,email",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		s.log.Debug("Validation errors at login", "errors", errs)
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	userID, err := s.auth.Register(ctx, req.GetUsername(), req.GetPassword(), req.GetEmail())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			errorMsg, err := json.Marshal(map[string]string{"email": err.Error()})
			if err != nil {
				slog.Error("Failed to marshal error message", "error", err)
				return nil, status.Error(codes.Internal, "failed to register")
			}
			return nil, status.Error(codes.AlreadyExists, string(errorMsg))
		}
		return nil, status.Error(codes.Internal, "failed to register")
	}
	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) GetOrCreateApp(ctx context.Context, req *ssov1.GetOrCreateAppRequest) (*ssov1.GetOrCreateAppResponse, error) {
	validationRules := map[string]string{
		"Name": "required,max=70",
		"Description": "required,max=300",
		"Secret": "required,min=12,max=64",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		s.log.Debug("Validation errors at login", "errors", errs)
		return nil, status.Error(codes.InvalidArgument, errs)
	}

	appID, created, err := s.auth.GetOrCreateApp(ctx, &models.App{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Secret:      req.GetSecret(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get or create app")
	}

	return &ssov1.GetOrCreateAppResponse{
		Id: appID,
		Created: created,
	}, nil
}