package auth

import (
	"context"
	"errors"

	"sso.service/internal/domain/models"
	"sso.service/internal/storage"
)

type GetAppParams struct {
	AppID   int32
	AppName string
}

type appsModel interface {
	Get(ctx context.Context, params GetAppParams) (*models.App, error)
	Create(ctx context.Context, app *models.App) (int64, error)
}

type AppIDAndIsCreated struct {
	AppID int64
	IsCreated bool
}

func (a *Auth) GetOrCreateApp(
	ctx context.Context,
	app *models.App,
) (*AppIDAndIsCreated, error) {
	const op = "auth.GetOrCreateApp"
	log := a.log.With("operation", op)
	id, err := a.appsModel.Create(ctx, app)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			log.Warn("App already exists", "name", app.Name)
			app, err := a.appsModel.Get(ctx, GetAppParams{AppName: app.Name})
			if err != nil {
				log.Error("Error getting app", "msg", err.Error())
				return nil, err
			}
			return &AppIDAndIsCreated{AppID: app.ID, IsCreated: false}, nil
		}
		log.Error("Error creating app", "msg", err.Error())
		return nil, err
	}
	log.Info("App saved", "id", id)
	return &AppIDAndIsCreated{AppID: id, IsCreated: true}, nil
}
