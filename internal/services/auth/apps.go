package auth

import (
	"context"
	"errors"

	"sso.service/internal/entity"
	"sso.service/internal/storage"
)

type GetAppOptionsDTO struct {
	AppID   int32
	AppName string
}

type appsRepo interface {
	Get(ctx context.Context, params GetAppOptionsDTO) (*entity.App, error)
	Create(ctx context.Context, app *entity.App) (int64, error)
}

type GetOrCreateAppDTO struct {
	AppID     int64
	IsCreated bool
}

func (a *AuthService) GetOrCreateApp(
	ctx context.Context,
	app *entity.App,
) (*GetOrCreateAppDTO, error) {
	const op = "auth.GetOrCreateApp"
	log := a.log.With("operation", op)
	id, err := a.appsRepo.Create(ctx, app)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			log.Warn("App already exists", "name", app.Name)
			app, err := a.appsRepo.Get(ctx, GetAppOptionsDTO{AppName: app.Name})
			if err != nil {
				log.Error("Error getting app", "msg", err.Error())
				return nil, err
			}
			return &GetOrCreateAppDTO{AppID: app.ID, IsCreated: false}, nil
		}
		log.Error("Error creating app", "msg", err.Error())
		return nil, err
	}
	log.Info("App saved", "id", id)
	return &GetOrCreateAppDTO{AppID: id, IsCreated: true}, nil
}
