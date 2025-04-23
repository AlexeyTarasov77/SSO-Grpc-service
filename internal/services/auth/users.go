package auth

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"sso.service/internal/entity"
	"sso.service/internal/services/dtos"
	"sso.service/internal/storage"
	jwtLib "sso.service/pkg/jwt"
)

type usersRepo interface {
	Get(ctx context.Context, params dtos.GetUserOptionsDTO) (*entity.User, error)
	GetForToken(ctx context.Context, tokenScope string, plainToken string) (*entity.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	Create(ctx context.Context, user *entity.User) (int64, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}

func (a *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
	appId int32,
) (*dtos.AuthTokens, error) {
	const op = "auth.Login"
	log := a.log.With("operation", op)
	isActive := new(bool)
	*isActive = true
	user, err := a.usersRepo.Get(ctx, dtos.GetUserOptionsDTO{Email: email, IsActive: isActive})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("Active user not found", "email", email)
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
	app, err := a.appsRepo.Get(ctx, dtos.GetAppOptionsDTO{AppID: appId})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
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
	return &dtos.AuthTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil

}

func (a *AuthService) Register(ctx context.Context, username string, plainPassword string, email string, appID int32) (*dtos.UserIDAndToken, error) {
	const op = "auth.Register"
	log := a.log.With("operation", op)
	user := entity.User{
		Username: username,
		Email:    email,
	}
	err := user.Password.Set(plainPassword)
	if err != nil {
		log.Error("Error setting password", "msg", err.Error())
		return nil, err
	}
	userID, err := a.usersRepo.Create(ctx, &user)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			log.Warn("User already exists", "email", email)
			return nil, ErrUserAlreadyExists
		}
		log.Error("Error saving user", "msg", err.Error())
		return nil, err
	}
	log.Info("User saved", "id", userID)
	app, err := a.appsRepo.Get(ctx, dtos.GetAppOptionsDTO{AppID: appID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("App not found", "app_id", appID)
			return nil, ErrAppNotFound
		}
		log.Error("Error getting app", "msg", err.Error())
		return nil, err
	}
	log.Info("Creating activation token", "userID", userID)
	tokenProvider := jwtLib.NewTokenProvider(app.Secret, a.cfg.TokenSigningAlg)
	claims := map[string]any{"uid": userID, "app_id": appID}
	token, err := tokenProvider.NewToken(a.cfg.ActivationTokenTTL, claims)
	if err != nil {
		log.Error("Error creating activation token", "msg", err.Error())
		return nil, err
	}
	log.Info("Activation token created", "token", token)
	return &dtos.UserIDAndToken{UserID: userID, Token: token}, nil
}

func (a *AuthService) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"
	log := a.log.With("operation", op, "user_id", userID)
	log.Info("Checking whether user is admin")
	isAdmin, err := a.usersRepo.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "user_id", userID)
			return false, ErrUserNotFound
		}
		a.log.Error("Error checking if user is admin", "msg", err.Error())
		return false, err
	}
	log.Info("Checked whether user is admin", "is_admin", isAdmin)
	return isAdmin, nil
}

func (a *AuthService) GetUser(ctx context.Context, params dtos.GetUserOptionsDTO) (*entity.User, error) {
	const op = "auth.GetUserByID"
	log := a.log.With("operation", op)
	user, err := a.usersRepo.Get(ctx, params)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "params", params)
			return nil, ErrUserNotFound
		}
		log.Error("Error getting user", "msg", err.Error())
		return nil, err
	}
	return user, nil
}

func (a *AuthService) ActivateUser(ctx context.Context, token string, appID int32) (*entity.User, error) {
	const op = "auth.ActivateUser"
	log := a.log.With("operation", op, "token", token, "appID", appID)
	app, err := a.appsRepo.Get(ctx, dtos.GetAppOptionsDTO{AppID: appID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("App not found", "app_id", appID)
			return nil, ErrAppNotFound
		}
		log.Error("Error getting app", "msg", err.Error())
		return nil, err
	}
	tokenProvider := jwtLib.NewTokenProvider(app.Secret, a.cfg.TokenSigningAlg)
	claims, err := tokenProvider.ParseClaimsFromToken(token)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenMalformed):
			log.Warn(err.Error(), "token", token)
			return nil, ErrInvalidToken
		default:
			log.Error("Error parsing token", "msg", err.Error())
			return nil, err
		}
	}
	userID := int64(claims["uid"].(float64))
	appIDFromToken := int32(claims["app_id"].(float64))
	if userID == 0 || appIDFromToken == 0 {
		log.Warn("Invalid activation token (missing uid or app_id)", "token", token)
		return nil, ErrInvalidToken
	}
	if appIDFromToken != appID {
		log.Warn("app_id mismatch", "appIDFromToken", appIDFromToken)
		return nil, ErrAppIdsMismatch
	}
	user, err := a.usersRepo.Get(ctx, dtos.GetUserOptionsDTO{ID: userID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("Invalid activation token (uid of unknown user)", "token", token)
			return nil, ErrInvalidToken
		}
		log.Error("Error getting user for token", "msg", err.Error())
		return nil, err
	}
	if user.IsActive {
		log.Warn("User already active", "email", user.Email)
		return nil, ErrUserAlreadyActivated
	}
	user.IsActive = true
	user, err = a.usersRepo.Update(ctx, user)
	if err != nil {
		log.Error("Error updating user", "msg", err.Error())
		return nil, err
	}
	return user, nil
}
