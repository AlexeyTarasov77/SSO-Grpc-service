package auth_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ssov1 "sso.service/api/proto/gen/v1"
	"sso.service/internal/entity"
	"sso.service/internal/services/dtos"
	"sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	validEmail := gofakeit.Email()
	validPassword := suite.FakePassword()
	validUsername := gofakeit.Username()
	testCases := []struct {
		name                   string
		req                    *ssov1.RegisterRequest
		expectedErr            bool
		expectedErrMsgContains string
		expectedCode           codes.Code
	}{
		{
			name: "valid",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: validPassword,
				Email:    validEmail,
				AppId:    suite.AppID,
			},
			expectedCode: codes.OK,
		},
		{
			name: "duplicate email",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: validPassword,
				Email:    validEmail,
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedErrMsgContains: "already exists",
			expectedCode:           codes.AlreadyExists,
		},
		{
			name: "invalid email",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: validPassword,
				Email:    "invalid email",
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedErrMsgContains: "email",
			expectedCode:           codes.InvalidArgument,
		},
		{
			name: "short password",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: "123",
				Email:    validEmail,
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedErrMsgContains: "password",
			expectedCode:           codes.InvalidArgument,
		},
		{
			name: "empty username",
			req: &ssov1.RegisterRequest{
				Username: "",
				Password: validPassword,
				Email:    validEmail,
				AppId:    suite.AppID,
			},
			expectedErr:            true,
			expectedErrMsgContains: "username",
			expectedCode:           codes.InvalidArgument,
		},
		{
			name: "no app id",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: validPassword,
				Email:    validEmail,
			},
			expectedErr:            true,
			expectedErrMsgContains: "app_id",
			expectedCode:           codes.InvalidArgument,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			respReg, err := st.AuthClient.Register(context.Background(), tc.req)
			respStatus := status.Code(err)
			t.Log("Actual code", respStatus, "Expected", tc.expectedCode, "req", tc.req)
			assert.Equal(t, tc.expectedCode, respStatus)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, respReg.GetUserId())
				if tc.expectedErrMsgContains != "" {
					assert.ErrorContains(t, err, tc.expectedErrMsgContains)
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, respReg.GetUserId())
			storage := st.NewTestStorage()
			entityObj := models.New(storage.DB)
			user, err := entityObj.User.Get(context.Background(), dtos.GetUserOptionsDTO{Email: validEmail})
			require.NoError(t, err)
			assert.Equal(t, respReg.GetUserId(), int64(user.ID))
			assert.Equal(t, validEmail, user.Email)
			assert.Equal(t, validUsername, user.Username)
			assert.Equal(t, entity.RoleUser, user.Role)
			assert.False(t, user.IsActive)
		})
	}
}
