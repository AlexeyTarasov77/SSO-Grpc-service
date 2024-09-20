package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type App struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Secret      string `db:"secret"`
}

type Role = string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID        int64     `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	Password  password  `db:"password" json:"-"`
	Role      Role      `db:"role" json:"role"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type password struct {
	Plaintext string
	Hash      []byte
}

func (p *password) Set(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.Plaintext = plain
	p.Hash = hash
	return nil
}

func (p *password) Matches(plain string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Scope     string
	Expiry    time.Time
}

func GenerateToken(userID int64, scope string, expiry time.Duration) (*Token, error) {
	token := &Token{
		UserID: userID,
		Scope:  scope,
		Expiry: time.Now().Add(expiry),
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}
