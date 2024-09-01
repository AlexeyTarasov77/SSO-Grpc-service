package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/domain/models"
	"sso.service/internal/storage"
)

type Storage struct {
	DB *pgxpool.Pool
}

func New(storagePath string) (*Storage, error) {
	dbpool, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		return nil, err
	}
	if err := dbpool.Ping(context.Background()); err != nil {
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
		user.Password,
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

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	// var user models.User
	row, _ := s.DB.Query(
		ctx,
		`SELECT id, username, email, password, role, is_active, created_at, updated_at FROM users
		WHERE email = $1 AND is_active = true`,
		email,
	)
	user, err := pgx.CollectOneRow(row, pgx.RowToStructByName[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}
		slog.Error("Failed to get user", "error", err)
		return models.User{}, err
	}
	return user, nil
}

func (s *Storage) App(ctx context.Context, id int32) (models.App, error) {
	row, _ := s.DB.Query(ctx, "SELECT * FROM apps WHERE id = $1", id)
	app, err := pgx.CollectOneRow(row, pgx.RowToStructByName[models.App])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}
		return models.App{}, err
	}
	return app, nil
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