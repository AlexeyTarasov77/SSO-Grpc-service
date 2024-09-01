package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"sso.service/internal/domain/models"
	"sso.service/internal/lib/jwt"
	"sso.service/internal/storage"
)

type UserSaver interface {
	SaveUser(ctx context.Context, user *models.User) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

// struct with params to get app by id or by name
type AppParams struct {
	AppID int32
	AppName string
}

type AppProvider interface {
	App(ctx context.Context, params AppParams) (models.App, error)
	CreateApp(ctx context.Context, app *models.App) (int64, error)
}

type Auth struct {
	log             *slog.Logger
	userSaver       UserSaver
	userProvider    UserProvider
	appProvider     AppProvider
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:             log,
		userSaver:       userSaver,
		userProvider:    userProvider,
		appProvider:     appProvider,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appId int32,
) (models.Tokens, error) {
	const op = "auth.Login"
	log := a.log.With("operation", op)
	emptyTokens := models.Tokens{}
	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", "email", email)
			return emptyTokens, ErrInvalidCredentials
		}
		log.Error("Error getting user", "msg", err)
		return emptyTokens, err
	}
	matches, err := user.Password.Matches(password)
	switch {
		case err != nil:
			log.Error("Error comparing password", "msg", err)
			return emptyTokens, err
		case !matches:
			log.Warn("Wrong password", "email", email)
			return emptyTokens, ErrInvalidCredentials
	}
	app, err := a.appProvider.App(ctx, AppParams{AppID: appId})
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("App not found", "app_id", appId)
			return emptyTokens, ErrInvalidCredentials
		}
		log.Error("Error getting app", "msg", err)
		return emptyTokens, err
	}
	accessToken, err := jwt.NewAccessToken(user, app, a.AccessTokenTTL)
	if err != nil {
		log.Error("Error creating access token", "msg", err)
		return emptyTokens, err
	}
	refreshToken, err := jwt.NewRefreshToken(app, a.RefreshTokenTTL)
	if err != nil {
		log.Error("Error creating refresh token", "msg", err)
		return emptyTokens, err
	}
	return models.Tokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil

}

func (a *Auth) Register(ctx context.Context, username string, plainPassword string, email string) (int64, error) {
	const op = "auth.Register"
	log := a.log.With("operation", op)
	user := models.User{
		Username: username,
		Email:    email,
	}
	err := user.Password.Set(plainPassword)
	if err != nil {
		log.Error("Error setting password", "msg", err)
		return 0, err
	}
	id, err := a.userSaver.SaveUser(ctx, &user)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Warn("User already exists", "email", email)
			return 0, ErrUserAlreadyExists
		}
		log.Error("Error saving user", "msg", err)
		return 0, err
	}
	log.Info("User saved", "id", id)
	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"
	log := a.log.With("operation", op, "user_id", userID)
	log.Info("Checking whether user is admin")
	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", "user_id", userID)
			return false, ErrUserNotFound
		}
		a.log.Error("Error checking if user is admin", "msg", err)
		return false, err
	}
	log.Info("Checked whether user is admin", "is_admin", isAdmin)
	return isAdmin, nil
}

func (a *Auth) GetOrCreateApp(
	ctx context.Context,
	app *models.App,
) (int64, bool, error) {
	const op = "auth.GetOrCreateApp"
	log := a.log.With("operation", op)
	id, err := a.appProvider.CreateApp(ctx, app)
	if err != nil {
		if errors.Is(err, storage.ErrAppAlreadyExists) {
			log.Warn("App already exists", "name", app.Name)
			app, err := a.appProvider.App(ctx, AppParams{AppName: app.Name})
			if err != nil {
				log.Error("Error getting app", "msg", err)
				return 0, false, err
			}
			return int64(app.ID), false, nil
		}
		log.Error("Error creating app", "msg", err)
		return 0, false, err
	}
	log.Info("App saved", "id", id)
	return id, true, nil
}