package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type App struct {
	ID     int `db:"id"`
	Name   string `db:"name"`
	Description string `db:"description"`
	Secret string `db:"secret"`
}

type Role string

const (
	RoleUser Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID        int `db:"id" json:"id"`
	Username  string `db:"username" json:"username"`
	Email     string `db:"email" json:"email"`
	Password  password `db:"password" json:"-"`
	Role      Role `db:"role" json:"role"`
	IsActive  bool `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type password struct {
	plaintext string
	hash      []byte
}

func (p *password) Set(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.plaintext = plain
	p.hash = hash
	return nil
}

func (p *password) Matches(plain string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}
