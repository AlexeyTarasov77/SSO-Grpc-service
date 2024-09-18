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
)

type AppModel struct {
	DB *pgxpool.Pool
}

func (a *AppModel) Create(ctx context.Context, app *models.App) (int64, error) {
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
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, storage.ErrAppAlreadyExists
		}
		return 0, err
	}
	return appID, nil
}

func (a *AppModel) Get(ctx context.Context, params auth.GetAppParams) (*models.App, error) {
	args := []any{params.AppID, params.AppName}
	row, _ := a.DB.Query(
		ctx,
		`SELECT id, name, coalesce(description, '') AS description, secret FROM apps WHERE (id = $1 OR $1 = 0) AND (name = $2 OR $2 = '')`,
		args...,
	)
	app, err := pgx.CollectOneRow(row, pgx.RowToStructByName[models.App])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrAppNotFound
		}
		return nil, err
	}
	return &app, nil
}
