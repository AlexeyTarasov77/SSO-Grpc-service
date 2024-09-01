package handlers

import (
	"testing"
	"time"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/tests/suite"
)

func TestLogin(t *testing.T) {
	t.Parallel()
	ctx, st := suite.New(t)
	validEmail := gofakeit.Email()
	validPassword := FakePassword()
	validUsername := gofakeit.Username()
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Username: validUsername,
		Password: validPassword,
		Email:    validEmail,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())
	testCases := []struct {
		name string
		req  *ssov1.LoginRequest
		expectedErr bool
		expectedErrMsg string
		expectedCode codes.Code
	} {
		{
			name: "valid",
			req: &ssov1.LoginRequest{
				Email:    validEmail,
				Password: validPassword,
				AppId:    appID,
			},
			expectedCode: codes.OK,
		},
		{
			name: "invalid email",
			req: &ssov1.LoginRequest{
				Email:    "invalid",
				Password: validPassword,
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.InvalidArgument,
			expectedErrMsg: "email",
		},
		{
			name: "short password",
			req: &ssov1.LoginRequest{
				Email:    validEmail,
				Password: "123",
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.InvalidArgument,
			expectedErrMsg: "password",
		},
		{
			name: "invalid app id",
			req: &ssov1.LoginRequest{
				Email:    validEmail,
				Password: validPassword,
				AppId:    emptyAppID,
			},
			expectedErr: true,
			expectedCode: codes.InvalidArgument,
			expectedErrMsg: "app_id",
		},
		{
			name: "Invalid credentials",
			req: &ssov1.LoginRequest{
				Email:    gofakeit.Email(),
				Password: validPassword,
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.InvalidArgument,
			expectedErrMsg: "invalid credentials",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			respLogin, err := st.AuthClient.Login(ctx, tc.req)
			loginTime := time.Now()
			assert.Equal(t, tc.expectedCode, status.Code(err))
			accessToken := respLogin.GetAccessToken()
			refreshToken := respLogin.GetRefreshToken()
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
				if tc.expectedErrMsg != "" {
					assert.ErrorContains(t, err, tc.expectedErrMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, accessToken)
			assert.NotEmpty(t, refreshToken)
			deltaSeconds := 3 * time.Second
			// verify access token
			tokenParsed, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
				return []byte(appSecret), nil
			})
			require.NoError(t, err)
			claims, ok := tokenParsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
			assert.Equal(t, validEmail, claims["email"].(string))
			assert.Equal(t, appID, int(claims["app_id"].(float64)))
			assert.InDelta(
				t,
				float64(loginTime.Add(st.Cfg.AccessTokenTTL).Unix()),
				claims["exp"].(float64),
				float64(deltaSeconds),
			)
			assert.True(t, tokenParsed.Valid)
			// verify refresh token
			tokenParsed, err = jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
				return []byte(appSecret), nil
			})
			require.NoError(t, err)
			claims, ok = tokenParsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.InDelta(
				t,
				float64(loginTime.Add(st.Cfg.RefreshTokenTTL).Unix()),
				claims["exp"].(float64),
				float64(deltaSeconds),
			)
			assert.True(t, tokenParsed.Valid)
		})
	}
}
