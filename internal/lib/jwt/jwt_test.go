package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test_secret"
const testSigningAlg = "HS256"
const testTokenExp = 10 * time.Minute

func TestNewToken(t *testing.T) {
	tokenPayload := map[string]any{"id": float64(1)}
	tokenProvider := NewTokenProvider(testSecret, testSigningAlg)
	token, err := tokenProvider.NewToken(testTokenExp, tokenPayload)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(testSecret), nil
	}, jwt.WithValidMethods([]string{tokenProvider.SigningAlg}))
	require.NoError(t, err)
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t,tokenPayload["id"], claims["id"])
	assert.Equal(t, float64(time.Now().Add(testTokenExp).Unix()), claims["exp"])
	assert.True(t, tokenParsed.Valid)
}


func TestParseClaimsFromToken(t *testing.T) {
	tokenProvider := NewTokenProvider(testSecret, testSigningAlg)
	tokenPayload := map[string]any{"id": float64(1)}
	token, err := tokenProvider.NewToken(testTokenExp, tokenPayload)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	claims, err := tokenProvider.ParseClaimsFromToken(token)
	require.NoError(t, err)
	assert.Equal(t, tokenPayload["id"], claims["id"])
}