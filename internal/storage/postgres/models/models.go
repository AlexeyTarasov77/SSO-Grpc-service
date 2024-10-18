package models

import "github.com/jackc/pgx/v5/pgxpool"

type Models struct {
	User *UserModel
	App *AppModel
	Permission *PermissionModel
}

func New(db *pgxpool.Pool) *Models {
	return &Models{
		User: &UserModel{DB: db},
		App: &AppModel{DB: db},
		Permission: &PermissionModel{DB: db},
	}
}