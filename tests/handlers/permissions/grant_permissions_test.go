package permissions_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ssov1 "sso.service/api/proto/gen/v1"
	"sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestGrantPermissions(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	storage := st.NewTestStorage()
	models := models.New(storage.DB)
	permCodes := []string{gofakeit.Username(), gofakeit.Username()}
	user := suite.CreateActiveTestUser(t, models.User)
	testCases := []struct {
		name              string
		req               *ssov1.GrantPermissionsRequest
		expectedCode      codes.Code
		expectedPermCodes []string
	}{
		{
			name: "valid with new codes",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          user.ID,
				PermissionCodes: permCodes,
			},
			expectedCode:      codes.OK,
			expectedPermCodes: permCodes,
		},
		{
			name: "valid with already existent codes",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          user.ID,
				PermissionCodes: permCodes,
			},
			expectedCode:      codes.OK,
			expectedPermCodes: []string{},
		},
		{
			name: "not found UserId",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          suite.NotFoundUserID,
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
		{
			name: "invalid user id",
			req: &ssov1.GrantPermissionsRequest{
				UserId:          -1,
				PermissionCodes: permCodes,
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
