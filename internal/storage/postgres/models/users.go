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

type UserModel struct {
	DB *pgxpool.Pool
}

func (u *UserModel) Create(ctx context.Context, user *models.User) (int64, error) {
	var userID int64
	err := u.DB.QueryRow(
		ctx,
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id",
		user.Username,
		user.Password.Hash,
		user.Email,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgres.UniqueViolationErrCode {
			return 0, storage.ErrUserAlreadyExists
		}
		return 0, err
	}
	return userID, nil
}

func (u *UserModel) Get(ctx context.Context, params auth.GetUserParams) (*models.User, error) {
	if params.IsActive == nil {
		params.IsActive = new(bool)
		*params.IsActive = true
	}
	args := []any{params.Email, params.ID, *params.IsActive}
	var user models.User
	err := u.DB.QueryRow(
		ctx,
		`SELECT id, username, email, role, is_active, created_at, updated_at FROM users
		WHERE (email = $1 OR $1 = '') AND (id = $2 OR $2 = 0) AND is_active = $3`,
		args...
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (u *UserModel) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	var role string
	transation, err := u.DB.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer transation.Commit(ctx)
	err = transation.QueryRow(ctx, "SELECT role FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
		return false, err
	}
	return role == string(models.RoleAdmin), nil
}