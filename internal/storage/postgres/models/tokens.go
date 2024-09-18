package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"sso.service/internal/domain/models"
	"sso.service/internal/storage"
)


type TokenModel struct {
	DB *pgxpool.Pool
}

func (t *TokenModel) Insert(ctx context.Context, token *models.Token) error {
	_, err := t.DB.Exec(
		ctx,
		"INSERT INTO tokens (hash, user_id, expiry, scope) VALUES ($1, $2, $3, $4)",
		token.Hash,
		token.UserID,
		token.Expiry,
		token.Scope,
	)
	if err != nil {
		return err
	}
	return nil
}

func (t *TokenModel) GenerateAndInsert(ctx context.Context, userID int64, scope string, expiry time.Duration) (*models.Token, error) {
	token, err := models.GenerateToken(userID, scope, expiry)
	if err != nil {
		return nil, err
	}
	err = t.Insert(ctx, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (t *TokenModel) DeleteAllForUserWithScope(ctx context.Context, userID int64, scope string) error {
	result, err := t.DB.Exec(ctx, "DELETE FROM tokens WHERE user_id = $1 AND scope = $2", userID, scope)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return storage.ErrTokenNotFound
	}
	return err
}