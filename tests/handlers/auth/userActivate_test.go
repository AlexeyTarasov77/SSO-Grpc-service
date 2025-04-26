package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ssov1 "sso.service/api/proto/gen/v1"
	models "sso.service/internal/storage/postgres/models"
	"sso.service/pkg/jwt"
	"sso.service/tests/suite"
)

func TestActivateUser(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	storage := st.NewTestStorage()
	userModel := models.New(storage.DB).User
	inactiveUser := suite.NewTestUser(t, false)
	suite.SaveTestUser(t, userModel, inactiveUser)
	activatedUser := suite.CreateActiveTestUser(t, userModel)
	tokenProvider := jwt.NewTokenProvider(suite.AppSecret, st.Cfg.TokenSigningAlg)
	validToken, err := tokenProvider.NewToken(st.Cfg.ActivationTokenTTL, map[string]any{"uid": inactiveUser.ID, "app_id": suite.AppID})
	require.NoError(t, err)
	expiredToken, err := tokenProvider.NewToken(time.Millisecond, map[string]any{"uid": inactiveUser.ID, "app_id": suite.AppID})
	require.NoError(t, err)
	activatedUserToken, err := tokenProvider.NewToken(st.Cfg.ActivationTokenTTL, map[string]any{"uid": activatedUser.ID, "app_id": suite.AppID})
	require.NoError(t, err)
	tokenWithNotFoundUser, err := tokenProvider.NewToken(st.Cfg.ActivationTokenTTL, map[string]any{"uid": 0, "app_id": suite.AppID})
	require.NoError(t, err)
	testCases := []struct {
		name         string
		req          *ssov1.ActivateUserRequest
		expectedCode codes.Code
		expectedUser *ssov1.User
	}{
		{
			name: "valid",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: validToken,
				AppId:           suite.AppID,
			},
			expectedCode: codes.OK,
			expectedUser: &ssov1.User{
				Id:       inactiveUser.ID,
				Username: inactiveUser.Username,
				Email:    inactiveUser.Email,
				Role:     inactiveUser.Role,
				IsActive: true,
			},
		},
		{
			name: "invalid token",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: "invalid_token",
				AppId:           suite.AppID,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "expired token",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: expiredToken,
				AppId:           suite.AppID,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid app-id",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: validToken,
				AppId:           0,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Already active user",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: activatedUserToken,
				AppId:           suite.AppID,
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "Unknown user from token",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: tokenWithNotFoundUser,
				AppId:           suite.AppID,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Not found app",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: validToken,
				AppId:           999999,
			},
			expectedCode: codes.NotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := st.AuthClient.ActivateUser(context.Background(), tc.req)
			respStatus := status.Code(err)
			t.Log("Actual code", respStatus, "Expected", tc.expectedCode)
			assert.Equal(t, tc.expectedCode, respStatus)
			if tc.expectedUser != nil && resp.GetUser() != nil {
				assert.Equal(t, tc.expectedUser.Id, resp.User.Id)
				assert.Equal(t, tc.expectedUser.Username, resp.User.Username)
				assert.Equal(t, tc.expectedUser.Email, resp.User.Email)
				assert.Equal(t, tc.expectedUser.Role, resp.User.Role)
				assert.Equal(t, tc.expectedUser.IsActive, resp.User.IsActive)
			}
		})
	}
}
