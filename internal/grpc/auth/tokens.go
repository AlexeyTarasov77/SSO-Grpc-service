package auth

import (
	"context"
	"errors"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/lib/validator"
	"sso.service/internal/services/auth"
)

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
