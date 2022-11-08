package users

import (
	"encoding/hex"

	"github.com/pkg/errors"
)

type User struct {
	Name         string
	PasswordHash []byte
	PasswordSalt []byte
	Permissions  map[string]struct{}
}

func NewUser(username, passwordHashHex, passwordSaltHex string, permissions map[string]struct{}) (*User, error) {
	if username == "" {
		return nil, errors.New("username must not be empty")
	}

	if len(passwordHashHex) != 64 {
		return nil, errors.New("password hash must be 64 (hex encoded scrypt hash) in length")
	}

	if len(passwordSaltHex) != 64 {
		return nil, errors.New("password salt must be 64 (hex encoded) in length")
	}

	var err error
	passwordHash, err := hex.DecodeString(passwordHashHex)
	if err != nil {
		return nil, errors.New("password hash must be hex encoded")
	}

	passwordSalt, err := hex.DecodeString(passwordSaltHex)
	if err != nil {
		return nil, errors.New("password salt must be hex encoded")
	}

	return &User{
		Name:         username,
		PasswordHash: passwordHash,
		PasswordSalt: passwordSalt,
		Permissions:  permissions,
	}, nil
}

// Clone returns a copy of a user.
func (u *User) Clone() *User {
	permissionsCopy := make(map[string]struct{}, len(u.Permissions))
	for k := range u.Permissions {
		permissionsCopy[k] = struct{}{}
	}

	return &User{
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		PasswordSalt: u.PasswordSalt,
		Permissions:  permissionsCopy,
	}
}

// PermissionsSlice returns the permissions of the user as a slice.
func (u *User) PermissionsSlice() []string {
	permissions := make([]string, 0, len(u.Permissions))

	for k := range u.Permissions {
		permissions = append(permissions, k)
	}

	return permissions
}
