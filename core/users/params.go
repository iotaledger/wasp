package users

import (
	"github.com/iotaledger/hive.go/core/app"
)

type User struct {
	Password    string   `default:"" usage:"the password of the user"`
	Permissions []string `default:"" usage:"the users permissions"`
}

type ParametersUsers struct {
	Users map[string]*User `noflag:"true" usage:"the list of accepted users"`
}

var ParamsUsers = &ParametersUsers{
	Users: map[string]*User{
		"wasp": {
			Password: "wasp",
			Permissions: []string{
				"dashboard",
				"api",
				"chain.read",
				"chain.write",
			},
		},
	},
}

var params = &app.ComponentParams{
	Params: map[string]any{
		"users": ParamsUsers,
	},
	Masked: nil,
}
