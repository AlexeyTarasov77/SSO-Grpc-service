package dtos

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

type GetUserOptionsDTO struct {
	Email    string
	ID       int64
	IsActive *bool // made pointer to support nil values
}

type UserIDAndToken struct {
	UserID int64
	Token  string
}
