package auth_test

import "github.com/brianvoe/gofakeit/v7"

const (
	appID = 1
	emptyAppID = 0
    appSecret = "test_secret"
	defaultPasswordLength = 10
	notFoundUserID = int64(999999999)
)

func FakePassword() string {
	return gofakeit.Password(true, true, true, true, true, defaultPasswordLength)
}