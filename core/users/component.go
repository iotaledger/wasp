package users

import (
	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/users"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "Users",
			Params:    params,
			Configure: configure,
		},
	}
}

var CoreComponent *app.CoreComponent

func configure() error {
	users.InitUsers(loadUsersFromConfiguration())

	return nil
}

func loadUsersFromConfiguration() map[string]*users.UserData {
	userMap := make(map[string]*users.UserData)
	for name, user := range ParamsUsers.Users {
		userMap[name] = &users.UserData{
			Username:    name,
			Password:    user.Password,
			Permissions: user.Permissions,
		}
	}

	if len(userMap) == 0 {
		// During the transition phase, create a default user when the config is empty.
		// This keeps the old authentication working.
		userMap["wasp"] = &users.UserData{
			Username:    "wasp",
			Password:    "wasp",
			Permissions: []string{"api", "dashboard"},
		}
	}
	return userMap
}
