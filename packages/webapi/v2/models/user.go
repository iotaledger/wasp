package models

type User struct {
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
}

type AddUserRequest struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Permissions []string `json:"permissions"`
}

type UpdateUserPasswordRequest struct {
	Password string `json:"password"`
}

type UpdateUserPermissionsRequest struct {
	Permissions []string `json:"permissions"`
}
