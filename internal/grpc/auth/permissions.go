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

func (s *serverAPI) CreatePermission(ctx context.Context, req *ssov1.CreatePermissionRequest) (*ssov1.CreatePermissionResponse, error) {
	validationRules := map[string]string{"code": "required,min=6,max=32"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	permission, err := s.auth.CreatePermission(ctx, req.GetCode())
	if err != nil {
		if (errors.Is(err, auth.ErrPermissionAlreadyExists)) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create permission")
	}
	return &ssov1.CreatePermissionResponse{Created: true, Permission: &ssov1.Permission{
		Id:   permission.ID,
		Code: permission.Code,
	}}, nil
}