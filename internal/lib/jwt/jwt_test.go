package jwt

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sso.service/internal/domain/models"
)

func TestNewAccessToken(t *testing.T) {
	user := models.User{ID: 1, Email: gofakeit.Email()}
	app := models.App{ID: 1, Secret: "test_secret"}
	testCases := []struct {
		name string
		exp  time.Duration
		expectedErr bool
	} {
		{"Valid", 10 * time.Minute, false},
		{"Zero expires", 0, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := NewAccessToken(user, app, tc.exp)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, token)
			tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.Secret), nil
			})
			require.NoError(t, err)
			claims, ok := tokenParsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.Equal(t, user.ID, int(claims["uid"].(float64)))
			assert.Equal(t, user.Email, claims["email"])
			assert.Equal(t, app.ID, int(claims["app_id"].(float64)))
			assert.Equal(t, float64(time.Now().Add(10*time.Minute).Unix()), claims["exp"])
			assert.True(t, tokenParsed.Valid)
		})
	}
}


func TestNewRefreshToken(t *testing.T) {
	app := models.App{ID: 1, Secret: "test_secret"}
	testCases := []struct {
		name string
		exp  time.Duration
		expectedErr bool
	} {
		{"Valid", 10 * time.Minute, false},
		{"Zero expires", 0, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := NewRefreshToken(app, tc.exp)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, token)
			tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.Secret), nil
			})
			require.NoError(t, err)
			claims, ok := tokenParsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.Equal(t, float64(time.Now().Add(10*time.Minute).Unix()), claims["exp"])
			assert.True(t, tokenParsed.Valid)
		})
	}
}