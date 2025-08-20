// Package users provides user management and authentication functionality for the Wasp node.
package users

import (
	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/v2/packages/authentication/shared/permissions"
)

const (
	CfgUsers = "users.users"
)

type User struct {
	PasswordHash string   `default:"0000000000000000000000000000000000000000000000000000000000000000" usage:"the auth password+salt as a scrypt hash"`
	PasswordSalt string   `default:"0000000000000000000000000000000000000000000000000000000000000000" usage:"the auth salt used for hashing the password"`
	Permissions  []string `default:"" usage:"permissions of the user"`
}

// PermissionsMap returns the permissions of the user as a map.
func (u *User) PermissionsMap() map[string]struct{} {
	permissionsMap := make(map[string]struct{})
	for _, v := range u.Permissions {
		permissionsMap[v] = struct{}{}
	}

	return permissionsMap
}

type ParametersUsers struct {
	Users map[string]*User `noflag:"true" usage:"the list of accepted users"`
}

var ParamsUsers = &ParametersUsers{
	Users: map[string]*User{
		"wasp": {
			PasswordHash: "c34ec258dd87938c9c19228f7062ef17c847ed9fac4f9a284ccfebdbff08e3f9",
			PasswordSalt: "db32d4f152a3dadd81cd9b71074a4ea3346dbf8ff1998d33e9452091fff6f503",
			Permissions: []string{
				permissions.API,
				permissions.Write,
			},
		},
	},
}

var params = &app.ComponentParams{
	AdditionalParams: map[string]map[string]any{
		"usersConfig": {
			"users": ParamsUsers,
		},
	},
	Masked: nil,
}
