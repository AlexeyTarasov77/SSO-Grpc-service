package auth

import (
	"errors"
)

// type ErrPermissionsIgnored struct {
// 	ignoredCodes []string
// 	countAll     int
// }
//
// func (self ErrPermissionsIgnored) Error() string {
// 	return fmt.Sprintf(
// 		"%d of %d supplied permissions were not found or were already granted before: %s",
// 		len(self.ignoredCodes),
// 		self.countAll,
// 		strings.Join(self.ignoredCodes, ", "),
// 	)
// }

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound         = errors.New("user not found")
	ErrAppNotFound          = errors.New("app not found")
	ErrUserAlreadyExists    = errors.New("user with this email already exists")
	ErrUserAlreadyActivated = errors.New("user already activated")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrAppIdsMismatch       = errors.New("app ids mismatch")
)

