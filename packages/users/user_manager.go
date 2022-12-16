package users

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/core/basicauth"
	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/wasp/packages/util"
)

// UserManager handles the list of users that are stored in the user config.
// It calls a function if the list changed.
type UserManager struct {
	onChangeMap *onchangemap.OnChangeMap[string, util.ComparableString, *User]
}

// NewUserManager creates a new user manager.
func NewUserManager(storeCallback func([]*User) error) *UserManager {
	return &UserManager{
		onChangeMap: onchangemap.NewOnChangeMap(
			onchangemap.WithChangedCallback[string, util.ComparableString](storeCallback),
		),
	}
}

func (m *UserManager) EnableStoreOnChange() {
	m.onChangeMap.CallbacksEnabled(true)
}

// Users returns a copy of all known users.
func (m *UserManager) Users() map[string]*User {
	return m.onChangeMap.All()
}

// User returns a copy of a user.
func (m *UserManager) User(name string) (*User, error) {
	return m.onChangeMap.Get(util.ComparableString(name))
}

// AddUser adds a user to the user manager.
func (m *UserManager) AddUser(user *User) error {
	return m.onChangeMap.Add(user)
}

// ModifyUser modifies a user in the user manager.
func (m *UserManager) ModifyUser(user *User) error {
	_, err := m.onChangeMap.Modify(user.ID(), func(item *User) bool {
		*item = *user
		return true
	})
	return err
}

// ChangeUserPassword changes the password of a user.
func (m *UserManager) ChangeUserPassword(name string, passwordHash, passwordSalt []byte) error {
	user, err := m.User(name)
	if err != nil {
		return fmt.Errorf("unable to change password for user \"%s\": user does not exist", name)
	}

	user.PasswordHash = passwordHash
	user.PasswordSalt = passwordSalt

	if err := m.ModifyUser(user); err != nil {
		return fmt.Errorf("unable to change password for user \"%s\": %w", name, err)
	}

	return nil
}

// ChangeUserPermissions changes the permissions of a user.
func (m *UserManager) ChangeUserPermissions(name string, permissions map[string]struct{}) error {
	user, err := m.User(name)
	if err != nil {
		return fmt.Errorf("unable to change permissions for user \"%s\": user does not exist", name)
	}

	user.Permissions = permissions

	if err := m.ModifyUser(user); err != nil {
		return fmt.Errorf("unable to change permissions for user \"%s\": %w", name, err)
	}

	return nil
}

// RemoveUser removes a user from the user manager.
func (m *UserManager) RemoveUser(name string) error {
	return m.onChangeMap.Delete(util.ComparableString(name))
}

// DerivePasswordKey derives a password key by hashing the given password with a salt.
func DerivePasswordKey(password string, passwordSaltHex ...string) ([]byte, []byte, error) {
	if password == "" {
		return []byte{}, []byte{}, errors.New("password must not be empty")
	}

	var err error
	var passwordSaltBytes []byte
	if len(passwordSaltHex) > 0 {
		// salt was given
		if len(passwordSaltHex[0]) != 64 {
			return []byte{}, []byte{}, errors.New("the given salt must be 64 (hex encoded) in length")
		}

		passwordSaltBytes, err = hex.DecodeString(passwordSaltHex[0])
		if err != nil {
			return []byte{}, []byte{}, fmt.Errorf("parsing given salt failed: %w", err)
		}
	} else {
		passwordSaltBytes, err = basicauth.SaltGenerator(32)
		if err != nil {
			return []byte{}, []byte{}, fmt.Errorf("generating random salt failed: %w", err)
		}
	}

	passwordKeyBytes, err := basicauth.DerivePasswordKey([]byte(password), passwordSaltBytes)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("deriving password key failed: %w", err)
	}

	return passwordKeyBytes, passwordSaltBytes, nil
}
