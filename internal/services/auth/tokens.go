package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"sso.service/internal/services/dtos"
	"sso.service/internal/storage"
	jwtLib "sso.service/pkg/jwt"
)

func (a *AuthService) RenewAccessToken(ctx context.Context, refreshToken string, appId int32) (string, error) {
	const op = "auth.GetAccessToken"
	log := a.log.With("operation", op)
	app, err := a.appsRepo.Get(ctx, dtos.GetAppOptionsDTO{AppID: appId})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
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
	user, err := a.usersRepo.Get(ctx, dtos.GetUserOptionsDTO{ID: userID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "user_id", userID)
			return "", ErrUserNotFound
		}
		log.Error("Error getting user", "msg", err.Error())
		return "", err
	}
	return tokenProvider.NewToken(a.cfg.AccessTokenTTL, map[string]any{"uid": user.ID, "app_id": app.ID})
}

func (a *AuthService) NewActivationToken(ctx context.Context, email string, appID int32) (string, error) {
	const op = "auth.NewActivationToken"
	log := a.log.With("operation", op)
	email = strings.Trim(email, " ")
	user, err := a.GetUser(ctx, dtos.GetUserOptionsDTO{Email: email})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("User not found", "email", email)
			return "", ErrUserNotFound
		}
		return "", err
	}
	if user.IsActive {
		log.Warn("User already active", "email", email)
		return "", ErrUserAlreadyActivated
	}
	app, err := a.appsRepo.Get(ctx, dtos.GetAppOptionsDTO{AppID: appID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("App not found", "app_id", appID)
			return "", ErrAppNotFound
		}
		log.Error("Error getting app", "msg", err.Error())
		return "", err
	}
	tokenProvider := jwtLib.NewTokenProvider(app.Secret, a.cfg.TokenSigningAlg)
	token, err := tokenProvider.NewToken(a.cfg.ActivationTokenTTL, map[string]any{"uid": user.ID, "app_id": app.ID})
	if err != nil {
		log.Error("Error creating activation token", "msg", err.Error())
		return "", err
	}
	return token, nil
}

func (a *AuthService) VerifyToken(ctx context.Context, appID int32, token string) error {
	const op = "auth.VerifyToken"
	log := a.log.With("operation", op)
	app, err := a.appsRepo.Get(ctx, dtos.GetAppOptionsDTO{AppID: appID})
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Warn("App not found", "app_id", appID)
			return ErrAppNotFound
		}
		log.Error("Error getting app", "msg", err.Error())
		return err
	}
	tokenProvider := jwtLib.NewTokenProvider(app.Secret, a.cfg.TokenSigningAlg)
	_, err = tokenProvider.ParseClaimsFromToken(token)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenMalformed):
			log.Warn(err.Error(), "token", token)
			return ErrInvalidToken
		default:
			log.Error("Error parsing token", "msg", err.Error())
			return err
		}
	}
	return nil
}

