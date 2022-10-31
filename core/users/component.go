package users

import (
	"encoding/hex"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/configuration"
	"github.com/iotaledger/wasp/packages/users"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:    "Users",
			Params:  params,
			Provide: provide,
		},
	}
}

var CoreComponent *app.CoreComponent

func provide(c *dig.Container) error {
	type userManagerDeps struct {
		dig.In
		UsersConfig         *configuration.Configuration `name:"usersConfig"`
		UsersConfigFilePath *string                      `name:"usersConfigFilePath"`
	}

	if err := c.Provide(func(deps userManagerDeps) *users.UserManager {
		userManager := users.NewUserManager((func(users []*users.User) error {
			// store users from user manager to the config file
			cfgUsers := make(map[string]*User)

			for _, u := range users {
				cfgUsers[u.Name] = &User{
					PasswordHash: hex.EncodeToString(u.PasswordHash),
					PasswordSalt: hex.EncodeToString(u.PasswordSalt),
					Permissions:  u.PermissionsSlice(),
				}
			}

			if err := deps.UsersConfig.Set(CfgUsers, cfgUsers); err != nil {
				return err
			}

			return deps.UsersConfig.StoreFile(*deps.UsersConfigFilePath, 0o600)
		}))

		// add users from config file to the user manager
		for name, u := range ParamsUsers.Users {
			user, err := users.NewUser(name, u.PasswordHash, u.PasswordSalt, u.PermissionsMap())
			if err != nil {
				CoreComponent.LogPanicf("unable to add user to user manager %s: %s", name, err)
			}

			if err := userManager.AddUser(user); err != nil {
				CoreComponent.LogPanicf("unable to add user to user manager %s: %s", name, err)
			}
		}

		userManager.StoreOnChange(true)

		return userManager
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}
