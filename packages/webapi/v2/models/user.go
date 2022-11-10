package models

type User struct {
	Username    string
	Permissions []string
}

type AddUserRequest struct {
	Username    string
	Password    string
	Permissions []string
}

type UpdateUserPasswordRequest struct {
	Password string
}

type UpdateUserPermissionsRequest struct {
	Permissions []string
}
