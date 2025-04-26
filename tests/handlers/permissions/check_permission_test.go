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

func TestCheckPermission(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	storage := st.NewTestStorage()
	models := models.New(storage.DB)
	user := suite.CreateActiveTestUser(t, models.User)
	grantedPerms := []string{gofakeit.Username(), gofakeit.Username()}
	err := models.Permission.CreateManyIgnoreConflict(context.Background(), grantedPerms)
	require.NoError(t, err)
	_, err = models.Permission.AddForUserIgnoreConflict(context.Background(), user.ID, grantedPerms)
	require.NoError(t, err)
	testCases := []struct {
		name            string
		req             *ssov1.CheckPermissionRequest
		expectedCode    codes.Code
		expectedHasPerm bool
	}{
		{
			name: "valid",
			req: &ssov1.CheckPermissionRequest{
				UserId:         user.ID,
				PermissionCode: grantedPerms[0],
			},
			expectedCode:    codes.OK,
			expectedHasPerm: true,
		},
		{
			name: "valid not granted perm",
			req: &ssov1.CheckPermissionRequest{
				UserId:         user.ID,
				PermissionCode: gofakeit.Username(),
			},
			expectedCode:    codes.OK,
			expectedHasPerm: false,
		},
		{
			name: "not found UserId",
			req: &ssov1.CheckPermissionRequest{
				UserId:         suite.NotFoundUserID,
				PermissionCode: grantedPerms[1],
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "empty perm Code",
			req: &ssov1.CheckPermissionRequest{
				UserId:         suite.NotFoundUserID,
				PermissionCode: "",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid user_id",
			req: &ssov1.CheckPermissionRequest{
				UserId:         -1,
				PermissionCode: grantedPerms[0],
			},
			expectedCode: codes.InvalidArgument,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := st.PermissionsClient.CheckPermission(context.Background(), tc.req)
			respStatus := status.Code(err)
			t.Log("Actual code", respStatus, "Expected", tc.expectedCode)
			assert.Equal(t, tc.expectedCode, respStatus)
			if tc.expectedCode == codes.OK {
				t.Log("Actual HasPermission", resp.GetHasPermission(), "expected", tc.expectedHasPerm)
				assert.Equal(t, tc.expectedHasPerm, resp.GetHasPermission())
			}
		})
	}
}
