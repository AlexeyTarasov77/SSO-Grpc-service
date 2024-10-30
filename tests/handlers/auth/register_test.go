package auth_test

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
		expectedErrMsgContains string
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
			expectedErrMsgContains: "already exists",
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
			expectedErrMsgContains: "email",
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
			expectedErrMsgContains: "password",
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
			expectedErrMsgContains: "username",
			expectedCode: codes.InvalidArgument,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			respReg, err := st.AuthClient.Register(ctx, tc.req)
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
			modelsObj := dbModels.New(storage.DB)
			user, err := modelsObj.User.Get(ctx, auth.GetUserParams{Email: validEmail})
			require.NoError(t, err)
			assert.Equal(t, respReg.GetUserId(), int64(user.ID))
			assert.Equal(t, validEmail, user.Email)
			assert.Equal(t, validUsername, user.Username)
			assert.Equal(t, models.RoleUser, user.Role)
			assert.False(t, user.IsActive)
		})
	}
}