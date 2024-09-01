package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"sso.service/internal/domain/models"
)

func NewAccessToken(user models.User, app models.App, expires time.Duration) (string, error) {
	if expires <= 0 {
		return "", errors.New("expires must be > 0")
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(expires).Unix()
	claims["app_id"] = app.ID
	return token.SignedString([]byte(app.Secret))
}

func NewRefreshToken(app models.App, expires time.Duration) (string, error) {
	if expires <= 0 {
		return "", errors.New("expires must be > 0")
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(expires).Unix()
	return token.SignedString([]byte(app.Secret))
}
