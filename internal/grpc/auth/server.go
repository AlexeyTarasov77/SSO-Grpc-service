package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/domain/models"
	"sso.service/internal/lib/validator"
	"sso.service/internal/services/auth"
)

type Auth interface {
	Login(ctx context.Context, username string, password string, appId int32) (*models.AuthTokens, error)
	Register(ctx context.Context, username string, password string, email string, appId int32) (*auth.UserIDAndToken, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	GetOrCreateApp(ctx context.Context, app *models.App) (*auth.AppIDAndIsCreated, error)
	RenewAccessToken(ctx context.Context, refreshToken string, appId int32) (string, error)
	GetUser(ctx context.Context, params auth.GetUserParams) (*models.User, error)
	ActivateUser(ctx context.Context, token string,  appID int32) (*models.User, error)
	NewActivationToken(ctx context.Context, email string, appID int32) (string, error)
	VerifyToken(ctx context.Context, appID int32, token string) error
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
		"Email":    "required,email",
		"Password": "required,min=8",
		"AppId":    "required,gt=0",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
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
	validationRules := map[string]string{"userId": "required,gt=0"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to check whether user is admin")
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
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	data, err := s.auth.Register(ctx, req.GetUsername(), req.GetPassword(), req.GetEmail(), req.GetAppId())
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
		UserId:          data.UserID,
		ActivationToken: data.Token,
	}, nil
}

func (s *serverAPI) GetOrCreateApp(ctx context.Context, req *ssov1.GetOrCreateAppRequest) (*ssov1.GetOrCreateAppResponse, error) {
	validationRules := map[string]string{
		"Name":        "required,max=70",
		"Description": "required,max=300",
		"Secret":      "required,min=12,max=64",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}

	data, err := s.auth.GetOrCreateApp(ctx, &models.App{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Secret:      req.GetSecret(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get or create app")
	}

	return &ssov1.GetOrCreateAppResponse{
		Id:      data.AppID,
		Created: data.IsCreated,
	}, nil
}

func (s *serverAPI) RenewAccessToken(ctx context.Context, req *ssov1.RenewAccessTokenRequest) (*ssov1.RenewAccessTokenResponse, error) {
	validationRules := map[string]string{
		"refreshToken": "required,min=10,max=64",
		"appId":        "required,gt=0",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	token, err := s.auth.RenewAccessToken(ctx, req.GetRefreshToken(), req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to refresh token")
	}
	return &ssov1.RenewAccessTokenResponse{
		AccessToken: token,
	}, nil
}

func (s *serverAPI) GetUser(ctx context.Context, req *ssov1.GetUserRequest) (*ssov1.GetUserResponse, error) {
	validationRules := map[string]string{
		"id": "omitempty,gt=0",
		"email": "omitempty,email",
		"is_active": "omitempty,boolean",
	}
	if req.GetId() == 0 && req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "either id or email must be provided")
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	s.log.Debug("Get user", "user_id", req.GetId(), "email", req.GetEmail(), "is_active", req.GetIsActive())
	is_active := req.GetIsActive()
	user, err := s.auth.GetUser(ctx, auth.GetUserParams{
		ID:       req.GetId(),
		Email:    req.GetEmail(),
		IsActive: &is_active,
	})
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	return &ssov1.GetUserResponse{
		User: &ssov1.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			IsActive: user.IsActive,
			CreatedAt:  user.CreatedAt.String(),
			UpdatedAt:  user.UpdatedAt.String(),
		},
	}, nil
}

func (s *serverAPI) NewActivationToken(ctx context.Context, req *ssov1.NewActivationTokenRequest) (*ssov1.NewActivationTokenResponse, error) {
	validationRules := map[string]string{"email": "required,email"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	token, err := s.auth.NewActivationToken(ctx, req.GetEmail(), req.GetAppId())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, auth.ErrUserAlreadyActivated):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create activation token")
		}
	}
	return &ssov1.NewActivationTokenResponse{
		ActivationToken: token,
	}, nil
}

func (s *serverAPI) ActivateUser(ctx context.Context, req *ssov1.ActivateUserRequest) (*ssov1.ActivateUserResponse, error) {
	validationRules := map[string]string{"activation_token": "required"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		s.log.Debug("Validation errors at login", "errors", errs)
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	user, err := s.auth.ActivateUser(ctx, req.GetActivationToken(), req.GetAppId())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, auth.ErrUserAlreadyActivated):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, auth.ErrInvalidToken):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to activate user")
		}
	}
	return &ssov1.ActivateUserResponse{
		User: &ssov1.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			IsActive: user.IsActive,
			CreatedAt:  user.CreatedAt.String(),
			UpdatedAt:  user.UpdatedAt.String(),
		},
	}, nil
}

func (s *serverAPI) VerifyToken(ctx context.Context, req *ssov1.VerifyTokenRequest) (*ssov1.VerifyTokenResponse, error) {
	validationRules := map[string]string{"token": "required"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		s.log.Debug("Validation errors at login", "errors", errs)
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	err := s.auth.VerifyToken(ctx, req.GetAppId(), req.GetToken())
	isTokenValid := true
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			isTokenValid = false
		case errors.Is(err, auth.ErrAppNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to verify token")
		}
	}
	return &ssov1.VerifyTokenResponse{IsValid: isTokenValid}, nil
}