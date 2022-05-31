package users

type UserData struct {
	Username    string
	Password    string
	Permissions []string
}

var users map[string]*UserData

func InitUsers(userList map[string]*UserData) {
	users = userList
}

// TODO: Maybe add a DB connection later on, including functionality to remove/edit users?

func All() map[string]*UserData {
	return users
}

func GetUserByName(name string) *UserData {
	return users[name]
}

func GetPermissionMap(name string) map[string]bool {
	permissionMap := make(map[string]bool)
	user := users[name]

	if user == nil {
		return permissionMap
	}

	for _, permission := range user.Permissions {
		permissionMap[permission] = true
	}

	return permissionMap
}
