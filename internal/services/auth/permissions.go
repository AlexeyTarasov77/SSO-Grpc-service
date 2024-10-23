package auth

import (
	"context"
	"errors"

	"sso.service/internal/domain/models"
	"sso.service/internal/storage"
)

type permissionsModel interface {
	Create(ctx context.Context, code string) (*models.Permission, error)
}

func (a *Auth) CreatePermission(ctx context.Context, code string) (*models.Permission, error) {
	perm, err := a.permissionsModel.Create(ctx, code)
	if err != nil {
		if (errors.Is(err, storage.ErrRecordAlreadyExists)) {
			return nil, ErrPermissionAlreadyExists
		}
		return nil, err
	}
	return perm, nil
}