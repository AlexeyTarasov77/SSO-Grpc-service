package permissions

import (
	"context"
	"errors"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/services/auth"
	"sso.service/internal/services/permissions"
	"sso.service/pkg/validator"
)

func (s *PermissionsServer) CheckPermission(ctx context.Context, req *ssov1.CheckPermissionRequest) (*ssov1.CheckPermissionResponse, error) {
	validationRules := map[string]string{"code": "required,min=6,max=32"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	hasPermission, err := s.service.CheckPermission(ctx, req.GetUserId(), req.GetPermissionCode())
	if err != nil {
		switch {
		case errors.Is(err, permissions.ErrPermissionNotFound) || errors.Is(err, auth.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to check permission")
	}
	return &ssov1.CheckPermissionResponse{HasPermission: hasPermission}, nil
}

func (s *PermissionsServer) GrantPermissions(ctx context.Context, req *ssov1.GrantPermissionsRequest) (*ssov1.GrantPermissionsResponse, error) {
	// TODO: only admin user can grant permissions to others
	validationRules := map[string]string{"userId": "required,gt=0", "permissionCodes": "required"}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}
	if len(req.GetPermissionCodes()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Permissions codes can't be emmtpy")
	}
	grantedPermissions, err := s.service.GrantPermissions(ctx, req.GetUserId(), req.GetPermissionCodes())
	if err != nil {
		if errors.Is(err, permissions.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to grant permission")
	}
	mappedPermissions := make([]*ssov1.Permission, len(grantedPermissions))
	for i, perm := range grantedPermissions {
		mappedPermissions[i] = &ssov1.Permission{
			Id:   perm.ID,
			Code: perm.Code,
		}
	}
	return &ssov1.GrantPermissionsResponse{GrantedPermissions: mappedPermissions}, nil
}
