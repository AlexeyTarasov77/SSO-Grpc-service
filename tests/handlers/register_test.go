package handlers

import (
	"testing"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/domain/models"
	"sso.service/internal/services/auth"
	dbModels "sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	ctx, st := suite.New(t)
	validEmail := gofakeit.Email()
	validPassword := FakePassword()
	validUsername := gofakeit.Username()
	testCases := []struct {
		name string
		req  *ssov1.RegisterRequest
		expectedErr bool
		expectedErrMsg string
		expectedCode codes.Code
	} {
		{
			name: "valid",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: validPassword,
				Email:    validEmail,
			},
			expectedCode: codes.OK,
		},
		{
			name: "duplicate email",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: validPassword,
				Email:    validEmail,
			},
			expectedErr: true,
			expectedErrMsg: "user already exists",
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "invalid email",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: validPassword,
				Email:    "invalid email",
			},
			expectedErr: true,
			expectedErrMsg: "email",
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "short password",
			req: &ssov1.RegisterRequest{
				Username: validUsername,
				Password: "123",
				Email:    validEmail,
			},
			expectedErr: true,
			expectedErrMsg: "password",
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty username",
			req: &ssov1.RegisterRequest{
				Username: "",
				Password: validPassword,
				Email:    validEmail,
			},
			expectedErr: true,
			expectedErrMsg: "username",
			expectedCode: codes.InvalidArgument,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			respReg, err := st.AuthClient.Register(ctx, tc.req)
			assert.Equal(t, tc.expectedCode, status.Code(err))
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, respReg.GetUserId())
				if tc.expectedErrMsg != "" {
					assert.ErrorContains(t, err, tc.expectedErrMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, respReg.GetUserId())
			storage := st.NewTestStorage(t)
			modelsObj := dbModels.New(storage.DB)
			user, err := modelsObj.User.Get(ctx, auth.GetUserParams{Email: validEmail})
			require.NoError(t, err)
			assert.Equal(t, respReg.GetUserId(), int64(user.ID))
			assert.Equal(t, validEmail, user.Email)
			assert.Equal(t, validUsername, user.Username)
			assert.Equal(t, models.RoleUser, user.Role)
			assert.True(t, user.IsActive)
		})
	}
}