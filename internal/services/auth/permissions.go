package auth

import (
	"context"
	"errors"

	"sso.service/internal/domain/models"
	"sso.service/internal/storage"
)

type PermissionGetParams struct {
	Code string
	ID int64
}

type permissionsModel interface {
	Create(ctx context.Context, code string) (*models.Permission, error)
	ExistsForUser(ctx context.Context, userID int64, code string) (bool, error)
	Get(ctx context.Context, params PermissionGetParams) (*models.Permission, error)
}

func (a *Auth) CreatePermission(ctx context.Context, code string) (*models.Permission, error) {
	const op = "auth.CreatePermission"
	log := a.log.With("operation", op, "code", code)
	perm, err := a.permissionsModel.Create(ctx, code)
	if err != nil {
		if (errors.Is(err, storage.ErrRecordAlreadyExists)) {
			log.Warn("Permission already exists")
			return nil, ErrPermissionAlreadyExists
		}
		log.Error("Failed to create permission", "msg", err.Error())
		return nil, err
	}
	return perm, nil
}

func (a *Auth) CheckPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	const op = "auth.CheckPermission"
	log := a.log.With("operation", op, "user_id", userID, "permission", permission)
	_, err := a.usersModel.Get(ctx, GetUserParams{ID: userID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "user_id", userID)
			return false, ErrUserNotFound
		}
		log.Error("Failed to get user", "msg", err.Error())
		return false, err
	}
	_, err = a.permissionsModel.Get(ctx, PermissionGetParams{Code: permission})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("Permission not found", "permission", permission)
			return false, ErrPermissionNotFound
		}
		log.Error("Failed to get permission", "msg", err.Error())
		return false, err
	}
	exists, err := a.permissionsModel.ExistsForUser(ctx, userID, permission)
	if err != nil {
		log.Error("Failed to check if permission exists", "msg", err.Error())
		return false, err
	}
	return exists, nil
}