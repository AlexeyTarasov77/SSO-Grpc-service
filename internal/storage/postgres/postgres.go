package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ( 
	UniqueViolationErrCode = "23505"
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