package suite

import "github.com/brianvoe/gofakeit/v7"

const (
	AppID          = 1
	EmptyAppID     = 0
	AppSecret      = "test-secret"
	NotFoundUserID = int64(999999999)
)

func FakePassword() string {
	return gofakeit.Password(true, true, true, true, true, 10)
}
