package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/domain/models"
	"sso.service/internal/services/auth"
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

func (p *PermissionModel) AddForUser(ctx context.Context, userID int64, codes ...string) error {
	const query = `
		INSERT INTO users_permissions 
		SELECT $1, p.id FROM permissions p WHERE p.code = ANY($2)`
	// pgtype.FlatArray[string](codes)
	args := []any{userID, codes}
	_, err := p.DB.Exec(ctx, query, args...) 
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			if pgxErr.Code == postgres.UniqueViolationErrCode {
                return storage.ErrRecordAlreadyExists
			} else if pgxErr.Code == postgres.ForeignKeyViolationErrCode {
				return storage.ErrRecordNotFound
			}
		}
        return err
	}
	return nil
}

func (p *PermissionModel) Get(ctx context.Context, params auth.PermissionGetParams) (*models.Permission, error) {
	var permission models.Permission
	const query = `
		SELECT * FROM permissions WHERE (id = $1 OR $1 = 0) AND (code = $2 OR $2 = '')`
	args := []any{params.ID, params.Code}
	err := p.DB.QueryRow(ctx, query, args...).Scan(&permission.ID, &permission.Code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}
	return &permission, nil
}

func (p *PermissionModel) ExistsForUser(ctx context.Context, userID int64, code string) (bool, error) {
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
		return false, err
	}
	return exists, nil
}