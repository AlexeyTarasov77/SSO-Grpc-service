package auth

import (
	"context"
	"encoding/json"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ssov1 "sso.service/api/proto/gen/v1"
	"sso.service/internal/services/auth"
	"sso.service/internal/services/dtos"
	"sso.service/pkg/validator"
)

func (s *AuthServer) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	validationRules := map[string]string{
		"Email":    "required,email",
		"Password": "required,min=8",
		"AppId":    "required,gt=0",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}

	tokens, err := s.service.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &ssov1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
func (s *AuthServer) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	validationRules := map[string]string{"UserId": "required,gt=0"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	isAdmin, err := s.service.IsAdmin(ctx, req.GetUserId())
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

func (s *AuthServer) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	validationRules := map[string]string{
		"Username": "required",
		"Password": "required,min=8",
		"Email":    "required,email",
		"AppId":    "required,gt=0",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	data, err := s.service.Register(ctx, req.GetUsername(), req.GetPassword(), req.GetEmail(), req.GetAppId())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserAlreadyExists):
			errorMsg, err := json.Marshal(map[string]string{"email": err.Error()})
			if err != nil {
				s.log.Error("Failed to marshal error message", "error", err)
				return nil, status.Error(codes.Internal, "failed to register")
			}
			return nil, status.Error(codes.AlreadyExists, string(errorMsg))
		case errors.Is(err, auth.ErrAppNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to register")
		}
	}
	return &ssov1.RegisterResponse{
		UserId:          data.UserID,
		ActivationToken: data.Token,
	}, nil
}

func (s *AuthServer) ActivateUser(ctx context.Context, req *ssov1.ActivateUserRequest) (*ssov1.ActivateUserResponse, error) {
	validationRules := map[string]string{"ActivationToken": "required", "AppId": "required,gt=0"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		s.log.Debug("Validation errors at login", "errors", errs)
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	user, err := s.service.ActivateUser(ctx, req.GetActivationToken(), req.GetAppId())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrAppNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, auth.ErrAppIdsMismatch):
			return nil, status.Error(codes.InvalidArgument, "Mismatch between app id provided in request and token")
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
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

func (s *AuthServer) GetUser(ctx context.Context, req *ssov1.GetUserRequest) (*ssov1.GetUserResponse, error) {
	validationRules := map[string]string{
		"Id":       "omitempty,gt=0",
		"Email":    "omitempty,email",
		"IsActive": "omitempty,boolean",
	}
	if req.GetId() == 0 && req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "either id or email must be provided")
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	isActive := req.GetIsActive()
	s.log.Debug("Get user", "user_id", req.GetId(), "email", req.GetEmail(), "is_active", isActive)
	user, err := s.service.GetUser(ctx, dtos.GetUserOptionsDTO{
		ID:       req.GetId(),
		Email:    req.GetEmail(),
		IsActive: &isActive,
	})
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	return &ssov1.GetUserResponse{
		User: &ssov1.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}
