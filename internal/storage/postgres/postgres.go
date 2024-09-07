package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/domain/models"
	"sso.service/internal/services/auth"
	"sso.service/internal/storage"
)

type Storage struct {
	DB *pgxpool.Pool
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	dbpool, err := pgxpool.New(ctx, storagePath)
	if err != nil {
		return nil, err
	}
	if err := dbpool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Storage{DB: dbpool}, nil
}

func (s *Storage) SaveUser(ctx context.Context, user *models.User) (int64, error) {
	var userID int64
	err := s.DB.QueryRow(
		ctx,
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id",
		user.Username,
		user.Password.Hash,
		user.Email,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, storage.ErrUserAlreadyExists
		}
		return 0, err
	}
	return userID, nil
}

func (s *Storage) User(ctx context.Context, params auth.GetUserParams) (models.User, error) {
	if params.IsActive == nil {
		params.IsActive = new(bool)
		*params.IsActive = true
	}
	args := []any{params.Email, params.ID, *params.IsActive}
	var user models.User
	err := s.DB.QueryRow(
		ctx,
		`SELECT id, username, email, password, role, is_active, created_at, updated_at FROM users
		WHERE (email = $1 OR $1 = '') AND (id = $2 OR $2 = 0) AND is_active = $3`,
		args...
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password.Hash, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}
		slog.Error("Failed to get user", "error", err)
		return models.User{}, err
	}
	return user, nil
}

func (s *Storage) App(ctx context.Context, params auth.AppParams) (models.App, error) {
	args := []any{params.AppID, params.AppName}
	row, _ := s.DB.Query(
		ctx,
		`SELECT id, name, coalesce(description, '') AS description, secret FROM apps WHERE (id = $1 OR $1 = 0) AND (name = $2 OR $2 = '')`,
		args...,
	)
	app, err := pgx.CollectOneRow(row, pgx.RowToStructByName[models.App])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}
		return models.App{}, err
	}
	return app, nil
}

func (s *Storage) CreateApp(ctx context.Context, app *models.App) (int64, error) {
	var appID int64
	err := s.DB.QueryRow(
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

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	var role string
	transation, err := s.DB.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer transation.Commit(ctx)
	err = transation.QueryRow(ctx, "SELECT role FROM users WHERE id = $1 FOR UPDATE", userID).Scan(role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
		return false, err
	}
	return role == string(models.RoleAdmin), nil
}