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
	dbentity "sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestIsAdmin(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	storage := st.NewTestStorage()
	userModel := dbentity.New(storage.DB).User
	userAdmin := entity.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Role:     entity.RoleAdmin,
		IsActive: true,
	}
	userAdmin.Password.Set(suite.FakePassword())
	validUserID, err := userModel.Create(context.Background(), &userAdmin)
	require.NoError(t, err)
	userNotAdmin := entity.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Role:     entity.RoleUser,
		IsActive: true,
	}
	userNotAdmin.Password.Set(suite.FakePassword())
	notAdminUserID, err := userModel.Create(context.Background(), &userNotAdmin)
	require.NoError(t, err)
	testCases := []struct {
		name            string
		req             *ssov1.IsAdminRequest
		expectedCode    codes.Code
		expectedIsAdmin bool
	}{
		{
			name: "valid",
			req: &ssov1.IsAdminRequest{
				UserId: validUserID,
			},
			expectedCode:    codes.OK,
			expectedIsAdmin: true,
		},
		{
			name: "invalid user id",
			req: &ssov1.IsAdminRequest{
				UserId: 0,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "not admin",
			req: &ssov1.IsAdminRequest{
				UserId: notAdminUserID,
			},
			expectedCode:    codes.OK,
			expectedIsAdmin: false,
		},
		{
			name: "not found user",
			req: &ssov1.IsAdminRequest{
				UserId: suite.NotFoundUserID,
			},
			expectedCode: codes.NotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp, err := st.AuthClient.IsAdmin(context.Background(), tc.req)
			respStatus := status.Code(err)
			t.Log("Actual code", respStatus, "Expected", tc.expectedCode, "userId", tc.req.GetUserId())
			assert.Equal(t, tc.expectedCode, respStatus)
			if tc.expectedCode == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tc.expectedIsAdmin, resp.GetIsAdmin())
			}
		})
	}
}
