package models

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/domain/models"
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
        return nil, err
    }
    return &permission, nil
}

func (p *PermissionModel) GetAllForUser(ctx context.Context, userID int64) (models.Permissions, error) {
	const query = `
		SELECT p.code FROM permissions p 
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
		var code string
		err = rows.Scan(&code)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}
	return permissions, nil
}