package models

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/entity"
	"sso.service/internal/services/dtos"
	"sso.service/internal/storage"
	"sso.service/internal/storage/postgres"
)

type UserModel struct {
	DB *pgxpool.Pool
}

func (u *UserModel) Create(ctx context.Context, user *entity.User) (int64, error) {
	if user.Role == "" {
		user.Role = entity.DefaultUserRole
	}
	var userID int64
	err := u.DB.QueryRow(
		ctx,
		"INSERT INTO users (username, password, email, is_active, role) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Username,
		user.Password.Hash,
		user.Email,
		user.IsActive,
		user.Role,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgres.UniqueViolationErrCode {
			return 0, storage.ErrRecordAlreadyExists
		}
		return 0, err
	}
	return userID, nil
}

func (u *UserModel) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	var updatedUser entity.User
	err := u.DB.QueryRow(
		ctx,
		`UPDATE users SET username = $1, password = $2, email = $3, role = $4, is_active = $5 WHERE id = $6 RETURNING id, username, email, role, is_active, created_at, updated_at`,
		user.Username,
		user.Password.Hash,
		user.Email,
		user.Role,
		user.IsActive,
		user.ID,
	).Scan(&updatedUser.ID, &updatedUser.Username, &updatedUser.Email, &updatedUser.Role, &updatedUser.IsActive, &updatedUser.CreatedAt, &updatedUser.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}
	return &updatedUser, nil
}

func (u *UserModel) Get(ctx context.Context, params dtos.GetUserOptionsDTO) (*entity.User, error) {
	args := []any{params.Email, params.ID}
	query := `SELECT id, username, password, email, role, is_active, created_at, updated_at FROM users
		WHERE (email = $1 OR $1 = '') AND (id = $2 OR $2 = 0)`
	if params.IsActive != nil {
		args = append(args, *params.IsActive)
		query += " AND is_active = $3"
	}
	var user entity.User
	err := u.DB.QueryRow(
		ctx,
		query,
		args...,
	).Scan(&user.ID, &user.Username, &user.Password.Hash, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (u *UserModel) GetForToken(ctx context.Context, tokenScope string, plainToken string) (*entity.User, error) {
	var user entity.User
	tokenHash := sha256.Sum256([]byte(plainToken))
	query := `
		SELECT u.id, u.username, u.password, u.email, u.role, u.is_active, u.created_at, u.updated_at FROM users u
		JOIN tokens t ON t.user_id = u.id
		WHERE t.hash = $1 AND t.scope = $2 AND t.expiry >= now()`
	args := []any{tokenHash[:], tokenScope}
	err := u.DB.QueryRow(
		ctx, query, args...,
	).Scan(&user.ID, &user.Username, &user.Password.Hash, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrRecordNotFound
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
			return false, storage.ErrRecordNotFound
		}
		return false, err
	}
	return role == string(entity.RoleAdmin), nil
}
