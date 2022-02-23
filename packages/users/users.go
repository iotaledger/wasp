package users

type User struct {
	Username string
	Password string
	Claims   []string
}

var users *[]User

func InitUsers(userList *[]User) {
	users = userList
}

// TODO: Maybe add a DB connection later on, including functionality to remove/edit users?

func All() *[]User {
	return users
}

func GetUserByName(name string) *User {
	for _, user := range *users {
		if user.Username == name {
			return &user
		}
	}

	return nil
}
