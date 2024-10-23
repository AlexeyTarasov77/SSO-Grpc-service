package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/domain/models"
	"sso.service/internal/storage"
	"sso.service/internal/storage/postgres"
)

type PermissionModel struct {
	DB *pgxpool.Pool
}

func (p *PermissionModel) Create(ctx context.Context, code string) (*models.Permission, error) {
    row, _ := p.DB.Query(
        ctx,
        "INSERT INTO permissions (code) VALUES ($1) RETURNING *",
        code,
    )
	permission, err := pgx.CollectOneRow(row, pgx.RowToStructByName[models.Permission])
    if err != nil {
		var err *pgconn.PgError
		if errors.As(err, &err) && err.Code == postgres.UniqueViolationErrCode {
			return nil, storage.ErrRecordAlreadyExists
		}
        return nil, err
    }
    return &permission, nil
}

func (p *PermissionModel) CheckForUser(ctx context.Context, userID int64, code string) (bool, error) {
	var exists bool
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM users_permissions WHERE user_id = $1 
			AND permission_id = (
				SELECT id FROM permissions WHERE code = $2
			)
		)`
	args := []any{userID, code}
	err := p.DB.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, storage.ErrRecordNotFound
		}
		return false, err
	}
	return exists, nil
}


func (p *PermissionModel) GetAllForUser(ctx context.Context, userID int64) (models.Permissions, error) {
	const query = `
		SELECT p.id, p.code FROM permissions p 
		JOIN users_permissions up ON p.id=up.permission_id
		JOIN users u ON u.id=up.user_id
		WHERE u.id=$1
	`
	args := []any{userID}
	rows, err := p.DB.Query(ctx, query, args...)
	if err != nil {
        return nil, err
    }
	defer rows.Close()
	permissions := make(models.Permissions, 0, rows.CommandTag().RowsAffected())
	for rows.Next() {
		var perm models.Permission
		err = rows.Scan(&perm.ID, &perm.Code)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}