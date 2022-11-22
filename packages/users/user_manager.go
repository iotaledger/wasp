package users

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/basicauth"
	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/wasp/packages/util"
)

// UserManager handles the list of users that are stored in the user config.
// It calls a function if the list changed.
type UserManager struct {
	storeOnChangeMap *onchangemap.OnChangeMap[string, util.ComparableString, *User]
}

// NewUserManager creates a new user manager.
func NewUserManager(storeCallback func([]*User) error) *UserManager {
	return &UserManager{
		storeOnChangeMap: onchangemap.NewOnChangeMap[string, util.ComparableString](storeCallback),
	}
}

func (m *UserManager) EnableStoreOnChange() {
	m.storeOnChangeMap.CallbackEnabled(true)
}

// Users returns a copy of all known users.
func (m *UserManager) Users() map[string]*User {
	return m.storeOnChangeMap.All()
}

// User returns a copy of a user.
func (m *UserManager) User(name string) (*User, error) {
	return m.storeOnChangeMap.Get(util.ComparableString(name))
}

// AddUser adds a user to the user manager.
func (m *UserManager) AddUser(user *User) error {
	return m.storeOnChangeMap.Add(user)
}

// ModifyUser modifies a user in the user manager.
func (m *UserManager) ModifyUser(user *User) error {
	_, err := m.storeOnChangeMap.Modify(user.ID(), func(item *User) bool {
		*item = *user
		return true
	})
	return err
}

// RemoveUser removes a user from the user manager.
func (m *UserManager) RemoveUser(name string) error {
	return m.storeOnChangeMap.Delete(util.ComparableString(name))
}

// DerivePasswordKey derives a password key by hashing the given password with a salt.
func DerivePasswordKey(password string, passwordSaltHex ...string) (string, string, error) {
	if password == "" {
		return "", "", errors.New("password must not be empty")
	}

	var err error
	var passwordSaltBytes []byte
	if len(passwordSaltHex) > 0 {
		// salt was given
		if len(passwordSaltHex[0]) != 64 {
			return "", "", errors.New("the given salt must be 64 (hex encoded) in length")
		}

		passwordSaltBytes, err = hex.DecodeString(passwordSaltHex[0])
		if err != nil {
			return "", "", fmt.Errorf("parsing given salt failed: %w", err)
		}
	} else {
		passwordSaltBytes, err = basicauth.SaltGenerator(32)
		if err != nil {
			return "", "", fmt.Errorf("generating random salt failed: %w", err)
		}
	}

	passwordKeyBytes, err := basicauth.DerivePasswordKey([]byte(password), passwordSaltBytes)
	if err != nil {
		return "", "", fmt.Errorf("deriving password key failed: %w", err)
	}

	return hex.EncodeToString(passwordKeyBytes), hex.EncodeToString(passwordSaltBytes), nil
}
