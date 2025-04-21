package auth

import (
	"context"
	"errors"

	"sso.service/internal/entity"
	"sso.service/internal/storage"
)

type GetPermissionOptionsDTO struct {
	Code string
	ID   int64
}

type permissionsRepo interface {
	Create(ctx context.Context, code string) (*entity.Permission, error)
	ExistsForUser(ctx context.Context, userID int64, code string) (bool, error)
	Get(ctx context.Context, params GetPermissionOptionsDTO) (*entity.Permission, error)
	AddForUser(ctx context.Context, userID int64, codes ...string) error
}

func (a *AuthService) CreatePermission(ctx context.Context, code string) (*entity.Permission, error) {
	const op = "auth.CreatePermission"
	log := a.log.With("operation", op, "code", code)
	perm, err := a.permissionsRepo.Create(ctx, code)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			log.Warn("Permission already exists")
			return nil, ErrPermissionAlreadyExists
		}
		log.Error("Failed to create permission", "msg", err.Error())
		return nil, err
	}
	return perm, nil
}

func (a *AuthService) CheckPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	const op = "auth.CheckPermission"
	log := a.log.With("operation", op, "user_id", userID, "permission", permission)
	_, err := a.usersRepo.Get(ctx, GetUserOptionsDTO{ID: userID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "user_id", userID)
			return false, ErrUserNotFound
		}
		log.Error("Failed to get user", "msg", err.Error())
		return false, err
	}
	_, err = a.permissionsRepo.Get(ctx, GetPermissionOptionsDTO{Code: permission})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("Permission not found", "permission", permission)
			return false, ErrPermissionNotFound
		}
		log.Error("Failed to get permission", "msg", err.Error())
		return false, err
	}
	exists, err := a.permissionsRepo.ExistsForUser(ctx, userID, permission)
	if err != nil {
		log.Error("Failed to check if permission exists", "msg", err.Error())
		return false, err
	}
	return exists, nil
}

func (a *AuthService) GrantPermissions(ctx context.Context, userID int64, permissionCodes ...string) error {
	const op = "auth.GrantPermission"
	log := a.log.With("operation", op, "user_id", userID, "permissionCodes", permissionCodes)
	err := a.permissionsRepo.AddForUser(ctx, userID, permissionCodes...)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			log.Warn("Some of the permissions already assigned to user")
			return ErrPermissionAlreadyExists
		} else if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "user_id", userID)
			return ErrUserNotFound
		}
		log.Error("Failed to grant permission", "msg", err.Error())
		return err
	}
	return nil
}

