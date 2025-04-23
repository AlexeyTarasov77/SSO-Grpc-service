package entity

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Role = string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"

	DefaultUserRole Role = RoleUser
)

type (
	User struct {
		ID          int64     `db:"id" json:"id"`
		Username    string    `db:"username" json:"username"`
		Email       string    `db:"email" json:"email"`
		Password    password  `db:"password" json:"-"`
		Role        Role      `db:"role" json:"role"`
		IsActive    bool      `db:"is_active" json:"is_active"`
		CreatedAt   time.Time `db:"created_at" json:"created_at"`
		UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
		Permissions []Permission
	}
	password struct {
		Plaintext string
		Hash      []byte
	}
)

func (u *User) HasPermission(permCode string) bool {
	for _, perm := range u.Permissions {
		if perm.Code == permCode {
			return true
		}
	}
	return false
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
