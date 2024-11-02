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
	dbModels "sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestIsAdmin(t *testing.T) {
	t.Parallel()
	ctx, st := suite.New(t)
	storage := st.NewTestStorage()
	userModel := dbModels.New(storage.DB).User
	userAdmin := models.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
        Role:     models.RoleAdmin,
        IsActive: true,
	}
	userAdmin.Password.Set(FakePassword())
	validUserID, err := userModel.Create(ctx, &userAdmin)
	require.NoError(t, err)
	userNotAdmin := models.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Role:     models.RoleUser,
		IsActive: true,
	}
	userNotAdmin.Password.Set(FakePassword())
	notAdminUserID, err := userModel.Create(ctx, &userNotAdmin)
	require.NoError(t, err)
	testCases := []struct {
		name string
		req  *ssov1.IsAdminRequest
		expectedCode codes.Code
		expectedIsAdmin bool
	} {
		{
			name: "valid",
			req: &ssov1.IsAdminRequest{
				UserId: validUserID,
			},
			expectedCode: codes.OK,
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
            expectedCode: codes.OK,
			expectedIsAdmin: false,
		},
		{
			name: "not found user",
			req: &ssov1.IsAdminRequest{
				UserId: notFoundUserID,
			},
			expectedCode: codes.NotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp, err := st.AuthClient.IsAdmin(ctx, tc.req)
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