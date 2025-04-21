package auth_test

import (
	"context"
	"testing"
	"time"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/entity"
	models "sso.service/internal/storage/postgres/models"
	"sso.service/pkg/jwt"
	"sso.service/tests/suite"
)

func TestActivateUser(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	storage := st.NewTestStorage()
	userModel := models.New(storage.DB).User
	validUser := entity.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Role:     entity.RoleUser,
		IsActive: false,
	}
	validUser.Password.Set(FakePassword())
	validUserID, err := userModel.Create(context.Background(), &validUser)
	require.NoError(t, err)
	validUser.ID = validUserID
	activatedUser := entity.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Role:     entity.RoleUser,
		IsActive: true,
	}
	activatedUser.Password.Set(FakePassword())
	activatedUserID, err := userModel.Create(context.Background(), &activatedUser)
	require.NoError(t, err)
	activatedUser.ID = activatedUserID
	tokenProvider := jwt.NewTokenProvider(appSecret, st.Cfg.TokenSigningAlg)
	validToken, err := tokenProvider.NewToken(st.Cfg.ActivationTokenTTL, map[string]any{"uid": validUser.ID, "app_id": appID})
	require.NoError(t, err)
	expiredToken, err := tokenProvider.NewToken(time.Millisecond, map[string]any{"uid": validUser.ID, "app_id": appID})
	require.NoError(t, err)
	activatedUserToken, err := tokenProvider.NewToken(st.Cfg.ActivationTokenTTL, map[string]any{"uid": activatedUser.ID, "app_id": appID})
	require.NoError(t, err)
	tokenWithNotFoundUser, err := tokenProvider.NewToken(st.Cfg.ActivationTokenTTL, map[string]any{"uid": 0, "app_id": appID})
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
				AppId:           appID,
			},
			expectedCode: codes.OK,
			expectedUser: &ssov1.User{
				Id:       validUser.ID,
				Username: validUser.Username,
				Email:    validUser.Email,
				Role:     validUser.Role,
				IsActive: true,
			},
		},
		{
			name: "invalid token",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: "invalid_token",
				AppId:           appID,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "expired token",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: expiredToken,
				AppId:           appID,
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
				AppId:           appID,
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "Unknown user from token",
			req: &ssov1.ActivateUserRequest{
				ActivationToken: tokenWithNotFoundUser,
				AppId:           appID,
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

