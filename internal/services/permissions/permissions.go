package permissions

import (
	"context"
	"errors"
	"fmt"

	"sso.service/internal/entity"
	"sso.service/internal/services/dtos"
	"sso.service/internal/storage"
)

type usersRepo interface {
	Get(ctx context.Context, params dtos.GetUserOptionsDTO) (*entity.User, error)
}

type permissionsRepo interface {
	ExistsForUser(ctx context.Context, userID int64, code string) (bool, error)
	Get(ctx context.Context, params dtos.GetPermissionOptionsDTO) (*entity.Permission, error)
	AddForUserIgnoreConflict(ctx context.Context, userID int64, codes []string) ([]int, error)
	FetchMany(ctx context.Context, options dtos.FetchManyPermissionsOptionsDTO) ([]entity.Permission, error)
	CreateManyIgnoreConflict(ctx context.Context, codes []string) error
}

func (a *PermissionsService) CheckPermission(ctx context.Context, userID int64, permCode string) (bool, error) {
	const op = "permissions.CheckPermission"
	log := a.log.With("operation", op, "user_id", userID, "permission", permCode)
	_, err := a.usersRepo.Get(ctx, dtos.GetUserOptionsDTO{ID: userID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "user_id", userID)
			return false, ErrUserNotFound
		}
		log.Error("Failed to get user", "msg", err.Error())
		return false, err
	}
	exists, err := a.permissionsRepo.ExistsForUser(ctx, userID, permCode)
	if err != nil {
		log.Error("Failed to check if permission exists", "msg", err.Error())
		return false, err
	}
	return exists, nil
}

func (a *PermissionsService) GrantPermissions(ctx context.Context, userID int64, permissionCodes []string) ([]entity.Permission, error) {
	const op = "permissions.GrantPermission"
	log := a.log.With("operation", op, "user_id", userID, "permissionCodes", permissionCodes)
	var grantedPermissions []entity.Permission
	if err := a.permissionsRepo.CreateManyIgnoreConflict(ctx, permissionCodes); err != nil {
		log.Error("Failed to create permissions", "msg", err.Error())
		return grantedPermissions, err
	}
	grantedPermissionIds, err := a.permissionsRepo.AddForUserIgnoreConflict(ctx, userID, permissionCodes)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "user_id", userID)
			return grantedPermissions, ErrUserNotFound
		}
		log.Error("Failed to grant permission", "msg", err.Error())
		return grantedPermissions, err
	}
	grantedPermissions, err = a.permissionsRepo.FetchMany(ctx, dtos.FetchManyPermissionsOptionsDTO{Ids: grantedPermissionIds})
	if err != nil {
		log.Error("Failed to fetch granted permissions", "msg", err.Error())
		return grantedPermissions, err
	}
	if len(grantedPermissionIds) != len(permissionCodes) {
		log.Info("Some of the permissions were already granted", "count", fmt.Sprintf("%d of %d", len(permissionCodes)-len(grantedPermissions), len(permissionCodes)))
	}
	return grantedPermissions, nil
}
