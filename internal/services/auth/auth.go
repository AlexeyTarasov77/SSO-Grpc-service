package auth

import (
	"context"
	"errors"
	"log/slog"

	"sso.service/internal/config"
	"sso.service/internal/domain/models"
	jwtLib "sso.service/internal/lib/jwt"
	"sso.service/internal/storage"
)

type UserSaver interface {
	SaveUser(ctx context.Context, user *models.User) (int64, error)
}

type GetUserParams struct {
	Email string
	ID    int
	IsActive *bool
}

type UserProvider interface {
	User(ctx context.Context, params GetUserParams) (models.User, error)
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
	cfg             *config.Config
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	cfg *config.Config,
) *Auth {
	return &Auth{
		log,
		userSaver,
		userProvider,
		appProvider,
		cfg,
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
	user, err := a.userProvider.User(ctx, GetUserParams{Email: email})
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", "email", email)
			return emptyTokens, ErrInvalidCredentials
		}
		log.Error("Error getting user", "msg", err.Error())
		return emptyTokens, err
	}
	matches, err := user.Password.Matches(password)
	switch {
		case err != nil:
			log.Error("Error comparing password", "msg", err.Error())
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
		log.Error("Error getting app", "msg", err.Error())
		return emptyTokens, err
	}
	tokenProvider := jwtLib.NewTokenProvider(app.Secret, a.cfg.TokenSigningAlg)
	tokenPayload := map[string]any{"uid": user.ID, "app_id": app.ID}
	accessToken, err := tokenProvider.NewToken(a.cfg.AccessTokenTTL, tokenPayload)
	if err != nil {
		log.Error("Error creating access token", "msg", err.Error())
		return emptyTokens, err
	}
	refreshToken, err := tokenProvider.NewToken(a.cfg.RefreshTokenTTL, tokenPayload)
	if err != nil {
		log.Error("Error creating refresh token", "msg", err.Error())
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
		log.Error("Error setting password", "msg", err.Error())
		return 0, err
	}
	id, err := a.userSaver.SaveUser(ctx, &user)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Warn("User already exists", "email", email)
			return 0, ErrUserAlreadyExists
		}
		log.Error("Error saving user", "msg", err.Error())
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
		a.log.Error("Error checking if user is admin", "msg", err.Error())
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
				log.Error("Error getting app", "msg", err.Error())
				return 0, false, err
			}
			return int64(app.ID), false, nil
		}
		log.Error("Error creating app", "msg", err.Error())
		return 0, false, err
	}
	log.Info("App saved", "id", id)
	return id, true, nil
}

func (a *Auth) GetAccessToken(ctx context.Context, refreshToken string, appId int32) (string, error) {
	const op = "auth.GetAccessToken"
	log := a.log.With("operation", op)
	app, err := a.appProvider.App(ctx, AppParams{AppID: appId})
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("App not found", "app_id", appId)
			return "", ErrAppNotFound
		}
		log.Error("Error getting app", "msg", err.Error())
		return "", err
	}
	tokenProvider := jwtLib.NewTokenProvider(app.Secret, a.cfg.TokenSigningAlg)
	claims, err := tokenProvider.ParseClaimsFromToken(refreshToken)
	if err != nil {
		log.Error("Error parsing refresh token", "msg", err.Error())
		return "", err
	}
	userID := claims["uid"].(float64)
	user, err := a.userProvider.User(ctx, GetUserParams{ID: int(userID)})
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", "user_id", int(userID))
			return "", ErrUserNotFound
		}
		log.Error("Error getting user", "msg", err.Error())
		return "", err
	}
	return tokenProvider.NewToken(a.cfg.AccessTokenTTL, map[string]any{"uid": user.ID, "app_id": app.ID})
}