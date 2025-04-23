package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/entity"
	"sso.service/internal/services/dtos"
	"sso.service/internal/storage"
	"sso.service/internal/storage/postgres"
)

type PermissionModel struct {
	DB *pgxpool.Pool
}

func (p *PermissionModel) AddForUserIgnoreConflict(ctx context.Context, userID int64, codes []string) ([]int, error) {
	const query = `INSERT INTO users_permissions AS up (user_id, permission_id)
		SELECT $1, p.id FROM permissions p WHERE p.code = ANY($2)
		ON CONFLICT DO NOTHING
		RETURNING up.permission_id
	`
	args := []any{userID, codes}
	var permissionIds []int
	rows, err := p.DB.Query(ctx, query, args...)
	if err != nil {
		return permissionIds, err
	}
	permissionIds, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (int, error) {
		var id int
		err := row.Scan(&id)
		return id, err
	})
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == postgres.ForeignKeyViolationErrCode {
			return permissionIds, storage.ErrRecordNotFound
		}
		return permissionIds, err
	}
	return permissionIds, nil
}

func (p *PermissionModel) CreateManyIgnoreConflict(ctx context.Context, codes []string) error {
	codes_len := len(codes)
	if codes_len == 0 {
		return nil
	}
	args := make([]any, codes_len)
	query := `INSERT INTO permissions (code) VALUES `
	for i, code := range codes {
		format := "($%d)"
		if i != codes_len-1 {
			format += ", "
		}
		query += fmt.Sprintf(format, i+1)
		args[i] = code
	}
	query += " ON CONFLICT DO NOTHING"
	_, err := p.DB.Exec(ctx, query, args...)
	return err
}

func (p *PermissionModel) FetchMany(ctx context.Context, options dtos.FetchManyPermissionsOptionsDTO) ([]entity.Permission, error) {
	const query = `
		SELECT * FROM permissions 
		WHERE (id = ANY ($1) OR $1 IS NULL) AND 
		(code = ANY ($2) OR $2 IS NULL)`
	rows, err := p.DB.Query(ctx, query, options.Ids, options.Codes)
	var permissions []entity.Permission
	if err != nil {
		return permissions, err
	}
	permissions, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.Permission])
	if err != nil {
		return permissions, err
	}
	return permissions, nil
}

func (p *PermissionModel) Get(ctx context.Context, params dtos.GetPermissionOptionsDTO) (*entity.Permission, error) {
	var permission entity.Permission
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
