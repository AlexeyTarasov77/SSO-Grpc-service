package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/entity"
	"sso.service/internal/services/dtos"
	"sso.service/internal/storage"
	"sso.service/internal/storage/postgres"
)

type AppModel struct {
	DB *pgxpool.Pool
}

func (a *AppModel) Create(ctx context.Context, app *entity.App) (int64, error) {
	var appID int64
	err := a.DB.QueryRow(
		ctx,
		"INSERT INTO apps (name, description, secret) VALUES ($1, $2, $3) RETURNING id",
		app.Name,
		app.Description,
		app.Secret,
	).Scan(&appID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgres.UniqueViolationErrCode {
			return 0, storage.ErrRecordAlreadyExists
		}
		return 0, err
	}
	return appID, nil
}

func (a *AppModel) Get(ctx context.Context, params dtos.GetAppOptionsDTO) (*entity.App, error) {
	args := []any{params.AppID, params.AppName}
	row, _ := a.DB.Query(
		ctx,
		`SELECT id, name, coalesce(description, '') AS description, secret FROM apps WHERE (id = $1 OR $1 = 0) AND (name = $2 OR $2 = '')`,
		args...,
	)
	app, err := pgx.CollectOneRow(row, pgx.RowToStructByName[entity.App])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}
	return &app, nil
}
