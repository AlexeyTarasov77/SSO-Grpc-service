package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"sso.service/internal/config"
	"sso.service/internal/domain/models"
	jwtLib "sso.service/internal/lib/jwt"
	"sso.service/internal/storage"
)

type GetUserParams struct {
	Email    string
	ID       int64
	IsActive *bool  // when nil will be evaluated to true
}

// struct with params to get app by id or by name
type GetAppParams struct {
	AppID   int32
	AppName string
}

type usersModel interface {
	Get(ctx context.Context, params GetUserParams) (*models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	Create(ctx context.Context, user *models.User) (int64, error)
}

type appsModel interface {
	Get(ctx context.Context, params GetAppParams) (*models.App, error)
	Create(ctx context.Context, app *models.App) (int64, error)
}

type tokensModel interface {
	GenerateAndInsert(ctx context.Context, userID int64, scope string, expiry time.Duration) (*models.Token, error)
}

type Auth struct {
	log          *slog.Logger
	usersModel usersModel
	appsModel  appsModel
	tokensModel tokensModel
	cfg          *config.Config
}

func New(
	log *slog.Logger,
	userModel usersModel,
	appsModel appsModel,
	tokensModel tokensModel,
	cfg *config.Config,
) *Auth {
	return &Auth{
		log,
		userModel,
		appsModel,
		tokensModel,
		cfg,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appId int32,
) (*models.AuthTokens, error) {
	const op = "auth.Login"
	log := a.log.With("operation", op)
	user, err := a.usersModel.Get(ctx, GetUserParams{Email: email})
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", "email", email)
			return nil, ErrInvalidCredentials
		}
		log.Error("Error getting user", "msg", err.Error())
		return nil, err
	}
	matches, err := user.Password.Matches(password)
	switch {
	case err != nil:
		log.Error("Error comparing password", "msg", err.Error())
		return nil, err
	case !matches:
		log.Warn("Wrong password", "email", email)
		return nil, ErrInvalidCredentials
	}
	app, err := a.appsModel.Get(ctx, GetAppParams{AppID: appId})
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("App not found", "app_id", appId)
			return nil, ErrInvalidCredentials
		}
		log.Error("Error getting app", "msg", err.Error())
		return nil, err
	}
	tokenProvider := jwtLib.NewTokenProvider(app.Secret, a.cfg.TokenSigningAlg)
	tokenPayload := map[string]any{"uid": user.ID, "app_id": app.ID}
	accessToken, err := tokenProvider.NewToken(a.cfg.AccessTokenTTL, tokenPayload)
	if err != nil {
		log.Error("Error creating access token", "msg", err.Error())
		return nil, err
	}
	refreshToken, err := tokenProvider.NewToken(a.cfg.RefreshTokenTTL, tokenPayload)
	if err != nil {
		log.Error("Error creating refresh token", "msg", err.Error())
		return nil, err
	}
	return &models.AuthTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil

}

type UserIDAndToken struct {
	UserID int64
	Token  *models.Token
}

func (a *Auth) Register(ctx context.Context, username string, plainPassword string, email string) (*UserIDAndToken, error) {
	const op = "auth.Register"
	log := a.log.With("operation", op)
	user := models.User{
		Username: username,
		Email:    email,
	}
	err := user.Password.Set(plainPassword)
	if err != nil {
		log.Error("Error setting password", "msg", err.Error())
		return nil, err
	}
	userID, err := a.usersModel.Create(ctx, &user)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Warn("User already exists", "email", email)
			return nil, ErrUserAlreadyExists
		}
		log.Error("Error saving user", "msg", err.Error())
		return nil, err
	}
	log.Info("User saved", "id", userID)
	log.Info("Creating activation token", "userID", userID)
	token, err := a.tokensModel.GenerateAndInsert(ctx, userID, models.ScopeActivation, a.cfg.ActivationTokenTTL)
	if err != nil {
		// TODO: Delete previously created user if can't create token
		log.Error("Error creating activation token", "msg", err.Error())
		return nil, err
	}
	log.Info("Activation token created", "token", token)
	return &UserIDAndToken{UserID: userID, Token: token}, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"
	log := a.log.With("operation", op, "user_id", userID)
	log.Info("Checking whether user is admin")
	isAdmin, err := a.usersModel.IsAdmin(ctx, userID)
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
		if errors.Is(err, storage.ErrAppAlreadyExists) {
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

func (a *Auth) RenewAccessToken(ctx context.Context, refreshToken string, appId int32) (string, error) {
	const op = "auth.GetAccessToken"
	log := a.log.With("operation", op)
	app, err := a.appsModel.Get(ctx, GetAppParams{AppID: appId})
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
	userID := int64(claims["uid"].(float64))
	user, err := a.usersModel.Get(ctx, GetUserParams{ID: userID})
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", "user_id", userID)
			return "", ErrUserNotFound
		}
		log.Error("Error getting user", "msg", err.Error())
		return "", err
	}
	return tokenProvider.NewToken(a.cfg.AccessTokenTTL, map[string]any{"uid": user.ID, "app_id": app.ID})
}

func (a *Auth) GetUser(ctx context.Context, params GetUserParams) (*models.User, error) {
	const op = "auth.GetUserByID"
	log := a.log.With("operation", op)
	user, err := a.usersModel.Get(ctx, params)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", "params", params)
			return nil, ErrUserNotFound
		}
		log.Error("Error getting user", "msg", err.Error())
		return nil, err
	}
	return user, nil
}