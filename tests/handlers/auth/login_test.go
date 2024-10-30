package auth_test

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
	"sso.service/internal/domain/models"
	dbModels "sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestLogin(t *testing.T) {
	t.Parallel()
	ctx, st := suite.New(t)
	userModel := dbModels.New(st.NewTestStorage().DB).User
	validUser := models.User{
		Username: gofakeit.Username(),
        Email:    gofakeit.Email(),
        IsActive: true,
	}
	validUser.Password.Set(FakePassword())
	validUserID, err := userModel.Create(ctx, &validUser)
	require.NoError(t, err)
	inActiveUser := models.User{
		Username: gofakeit.Username(),
        Email:    gofakeit.Email(),
        IsActive: false,
	}
	inActiveUser.Password.Set(FakePassword())
	_, err = userModel.Create(ctx, &inActiveUser)
	require.NoError(t, err)
	testCases := []struct {
		name string
		req  *ssov1.LoginRequest
		expectedErr bool
		expectedErrMsgContains string
		expectedCode codes.Code
	} {
		{
			name: "valid",
			req: &ssov1.LoginRequest{
				Email:    validUser.Email,
				Password: validUser.Password.Plaintext,
				AppId:    appID,
			},
			expectedCode: codes.OK,
		},
		{
			name: "invalid email",
			req: &ssov1.LoginRequest{
				Email:    "invalid",
				Password: validUser.Password.Plaintext,
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.InvalidArgument,
			expectedErrMsgContains: "email",
		},
		{
			name: "short password",
			req: &ssov1.LoginRequest{
				Email:    validUser.Email,
				Password: "123",
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.InvalidArgument,
			expectedErrMsgContains: "password",
		},
		{
			name: "invalid app-id",
			req: &ssov1.LoginRequest{
				Email:    validUser.Email,
				Password: validUser.Password.Plaintext,
				AppId:    emptyAppID,
			},
			expectedErr: true,
			expectedCode: codes.InvalidArgument,
			expectedErrMsgContains: "app_id",
		},
		{
			name: "Invalid credentials (unknown email)",
			req: &ssov1.LoginRequest{
				Email:    gofakeit.Email(),
				Password: validUser.Password.Plaintext,
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.Unauthenticated,
			expectedErrMsgContains: "invalid credentials",
		},
		{
			name: "Invalid credentials (wrong password)",
			req: &ssov1.LoginRequest{
				Email:    validUser.Email,
				Password: FakePassword(),
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.Unauthenticated,
			expectedErrMsgContains: "invalid credentials",
		},
		{
			name: "Invalid credentials (inactive user)",
			req: &ssov1.LoginRequest{
				Email:    inActiveUser.Email,
				Password: inActiveUser.Password.Plaintext,
				AppId:    appID,
			},
			expectedErr: true,
			expectedCode: codes.Unauthenticated,
			expectedErrMsgContains: "invalid credentials",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			respLogin, err := st.AuthClient.Login(ctx, tc.req)
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
				return []byte(appSecret), nil
			})
			require.NoError(t, err)
			claims, ok := tokenParsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.Equal(t, validUserID, int64(claims["uid"].(float64)))
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
