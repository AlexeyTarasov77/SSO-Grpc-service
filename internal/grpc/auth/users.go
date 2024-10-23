package auth

import (
	"context"
	"encoding/json"
	"errors"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/lib/validator"
	"sso.service/internal/services/auth"
)

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
				s.log.Error("Failed to marshal error message", "error", err)
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


func (s *serverAPI) GetUser(ctx context.Context, req *ssov1.GetUserRequest) (*ssov1.GetUserResponse, error) {
	validationRules := map[string]string{
		"id":        "omitempty,gt=0",
		"email":     "omitempty,email",
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