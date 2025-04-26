package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ssov1 "sso.service/api/proto/gen/v1"
	"sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestLogin(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	userModel := models.New(st.NewTestStorage().DB).User
	activatedUser := suite.CreateActiveTestUser(t, userModel)
	inactiveUser := suite.NewTestUser(t, false)
	suite.SaveTestUser(t, userModel, inactiveUser)
	testCases := []struct {
		name                   string
		req                    *ssov1.LoginRequest
		expectedErr            bool
		expectedErrMsgContains string
		expectedCode           codes.Code
	}{
		{
			name: "valid",
			req: &ssov1.LoginRequest{
				Email:    activatedUser.Email,
				Password: activatedUser.Password.Plaintext,
				AppId:    suite.AppID,
			},
			expectedCode: codes.OK,
		},
		{
			name: "invalid email",
			req: &ssov1.LoginRequest{
				Email:    "invalid",
				Password: activatedUser.Password.Plaintext,
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedCode:           codes.InvalidArgument,
			expectedErrMsgContains: "email",
		},
		{
			name: "short password",
			req: &ssov1.LoginRequest{
				Email:    activatedUser.Email,
				Password: "123",
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedCode:           codes.InvalidArgument,
			expectedErrMsgContains: "password",
		},
		{
			name: "invalid app-id",
			req: &ssov1.LoginRequest{
				Email:    activatedUser.Email,
				Password: activatedUser.Password.Plaintext,
				AppId:    suite.EmptyAppID,
			},
			expectedErr:            true,
			expectedCode:           codes.InvalidArgument,
			expectedErrMsgContains: "app_id",
		},
		{
			name: "Invalid credentials (unknown email)",
			req: &ssov1.LoginRequest{
				Email:    gofakeit.Email(),
				Password: activatedUser.Password.Plaintext,
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedCode:           codes.Unauthenticated,
			expectedErrMsgContains: "invalid credentials",
		},
		{
			name: "Invalid credentials (wrong password)",
			req: &ssov1.LoginRequest{
				Email:    activatedUser.Email,
				Password: suite.FakePassword(),
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedCode:           codes.Unauthenticated,
			expectedErrMsgContains: "invalid credentials",
		},
		{
			name: "Invalid credentials (inactive user)",
			req: &ssov1.LoginRequest{
				Email:    inactiveUser.Email,
				Password: inactiveUser.Password.Plaintext,
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedCode:           codes.Unauthenticated,
			expectedErrMsgContains: "invalid credentials",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			respLogin, err := st.AuthClient.Login(context.Background(), tc.req)
			loginTime := time.Now()
			t.Log("Expected status", tc.expectedCode, "got", status.Code(err), "user email", tc.req.Email, "name", tc.name)
			assert.Equal(t, tc.expectedCode, status.Code(err))
			accessToken := respLogin.GetAccessToken()
			refreshToken := respLogin.GetRefreshToken()
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
				if tc.expectedErrMsgContains != "" {
					assert.ErrorContains(t, err, tc.expectedErrMsgContains)
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, accessToken)
			assert.NotEmpty(t, refreshToken)
			deltaSeconds := 3 * time.Second
			// verify access token
			tokenParsed, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
				return []byte(suite.AppSecret), nil
			})
			require.NoError(t, err)
			claims, ok := tokenParsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.Equal(t, activatedUser.ID, int64(claims["uid"].(float64)))
			assert.Equal(t, suite.AppID, int(claims["app_id"].(float64)))
			assert.InDelta(
				t,
				float64(loginTime.Add(st.Cfg.AccessTokenTTL).Unix()),
				claims["exp"].(float64),
				float64(deltaSeconds),
			)
			assert.True(t, tokenParsed.Valid)
			// verify refresh token
			tokenParsed, err = jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
				return []byte(suite.AppSecret), nil
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
