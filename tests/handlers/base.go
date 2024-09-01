package handlers

import "github.com/brianvoe/gofakeit/v7"

const (
	appID = 1
	emptyAppID = 0
    appSecret = "test-secret"
	defaultPasswordLength = 10
)

func FakePassword() string {
	return gofakeit.Password(true, true, true, true, true, defaultPasswordLength)
}