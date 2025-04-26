package suite

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"sso.service/internal/entity"
	"sso.service/internal/storage/postgres/models"
)

const (
	AppID          = 1
	EmptyAppID     = 0
	AppSecret      = "test-secret"
	NotFoundUserID = int64(999999999)
)

func FakePassword() string {
	return gofakeit.Password(true, true, true, true, true, 10)
}

func NewTestUser(t *testing.T, isActive bool) *entity.User {
	user := entity.User{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		IsActive: isActive,
	}
	user.Password.Set(FakePassword())
	return &user
}

func SaveTestUser(t *testing.T, userModel *models.UserModel, user *entity.User) {
	validUserID, err := userModel.Create(context.Background(), user)
	require.NoError(t, err)
	user.ID = validUserID
}

func CreateActiveTestUser(t *testing.T, userModel *models.UserModel) *entity.User {
	user := NewTestUser(t, true)
	SaveTestUser(t, userModel, user)
	return user
}
