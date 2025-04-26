package auth_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ssov1 "sso.service/api/proto/gen/v1"
	"sso.service/internal/entity"
	"sso.service/internal/storage/postgres/models"
	"sso.service/tests/suite"
)

func TestGetUser(t *testing.T) {
	t.Parallel()
	st := suite.New(t)
	storage := st.NewTestStorage()
	userModel := models.New(storage.DB).User
	validUser := entity.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		IsActive: true,
	}
	validUser.Password.Set(suite.FakePassword())
	validUserID, err := userModel.Create(context.Background(), &validUser)
	validUser.ID = validUserID
	require.NoError(t, err)
	testCases := []struct {
		name         string
		req          *ssov1.GetUserRequest
		expectedCode codes.Code
		expectedUser *ssov1.User
	}{
		{
			name: "valid by id",
			req: &ssov1.GetUserRequest{
				Id:       validUserID,
				IsActive: true,
			},
			expectedCode: codes.OK,
			expectedUser: &ssov1.User{
				Id:       validUser.ID,
				Username: validUser.Username,
				Email:    validUser.Email,
				Role:     validUser.Role,
				IsActive: validUser.IsActive,
			},
		},
		{
			name: "valid by email",
			req: &ssov1.GetUserRequest{
				Email:    validUser.Email,
				IsActive: true,
			},
			expectedCode: codes.OK,
			expectedUser: &ssov1.User{
				Id:       validUser.ID,
				Username: validUser.Username,
				Email:    validUser.Email,
				Role:     validUser.Role,
				IsActive: validUser.IsActive,
			},
		},
		{
			name: "not found by id",
			req: &ssov1.GetUserRequest{
				Id:       suite.NotFoundUserID,
				IsActive: true,
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "not found by email",
			req: &ssov1.GetUserRequest{
				Email:    gofakeit.Email(),
				IsActive: true,
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "not found by IsActive",
			req: &ssov1.GetUserRequest{
				Id: validUserID,
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "invalid id",
			req: &ssov1.GetUserRequest{
				Id: -1,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid email",
			req: &ssov1.GetUserRequest{
				Email: "invalid",
			},
			expectedCode: codes.InvalidArgument,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp, err := st.AuthClient.GetUser(context.Background(), tc.req)
			respStatus := status.Code(err)
			t.Log("Actual code", respStatus, "Expected", tc.expectedCode)
			require.Equal(t, tc.expectedCode, respStatus)
			if resp.GetUser() != nil {
				t.Log("Actual user", resp.GetUser(), "Expected", tc.expectedUser)
				require.Equal(t, tc.expectedUser.Id, resp.GetUser().Id)
				require.Equal(t, tc.expectedUser.Username, resp.GetUser().Username)
				require.Equal(t, tc.expectedUser.Email, resp.GetUser().Email)
				require.Equal(t, tc.expectedUser.Role, resp.GetUser().Role)
				require.Equal(t, tc.expectedUser.IsActive, resp.GetUser().IsActive)
			}
		})
	}
}
