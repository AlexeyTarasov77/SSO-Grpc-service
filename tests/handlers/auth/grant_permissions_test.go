package auth_test

import (
	"context"
	"testing"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/entity"
	"sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestGrantPermissions(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	storage := st.NewTestStorage()
	models := models.New(storage.DB)
	precreatedPermCodes := []string{gofakeit.Username(), gofakeit.Username()}
	for _, code := range precreatedPermCodes {
		_, err := models.Permission.Create(context.Background(), code)
		require.NoError(t, err)
	}
	newPermCodes := []string{gofakeit.MovieName(), gofakeit.MovieName()}
	user := entity.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		IsActive: true,
	}
	user.Password.Set(FakePassword())
	validUserID, err := models.User.Create(context.Background(), &user)
	user.ID = validUserID
	require.NoError(t, err)
	testCases := []struct {
		name              string
		req               *ssov1.GrantPermissionsRequest
		expectedCode      codes.Code
		expectedPermCodes []string
	}{
		{
			name: "valid with precreated codes",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          user.ID,
				PermissionCodes: precreatedPermCodes,
			},
			expectedCode:      codes.OK,
			expectedPermCodes: precreatedPermCodes,
		},
		{
			name: "valid with new codes",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          user.ID,
				PermissionCodes: newPermCodes,
			},
			expectedCode:      codes.OK,
			expectedPermCodes: newPermCodes,
		},
		{
			name: "valid with already existent codes",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          user.ID,
				PermissionCodes: precreatedPermCodes,
			},
			expectedCode:      codes.OK,
			expectedPermCodes: []string{},
		},
		{
			name: "not found UserId",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          notFoundUserID,
				PermissionCodes: []string{gofakeit.Email(), gofakeit.Email()},
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "empty permissions",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          user.ID,
				PermissionCodes: []string{},
			},
			expectedCode: codes.InvalidArgument,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := st.PermissionsClient.GrantPermissions(context.Background(), tc.req)
			respStatus := status.Code(err)
			t.Log("Actual code", respStatus, "Expected", tc.expectedCode)
			require.Equal(t, tc.expectedCode, respStatus)
			if tc.expectedCode == codes.OK {
				t.Log("Actual permissions", resp.GetGrantedPermissions(), "expected", tc.expectedPermCodes)
				for _, code := range tc.expectedPermCodes {
					inResponse := false
					for _, perm := range resp.GetGrantedPermissions() {
						if perm.GetCode() == code {
							inResponse = true
							break
						}
					}
					assert.Equal(t, inResponse, true)
				}
			}
		})
	}
}
