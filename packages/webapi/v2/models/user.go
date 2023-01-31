package models

type User struct {
	Username    string   `json:"username" swagger:"required"`
	Permissions []string `json:"permissions" swagger:"required"`
}

type AddUserRequest struct {
	Username    string   `json:"username" swagger:"required"`
	Password    string   `json:"password" swagger:"required"`
	Permissions []string `json:"permissions" swagger:"required"`
}

type UpdateUserPasswordRequest struct {
	Password string `json:"password" swagger:"required"`
}

type UpdateUserPermissionsRequest struct {
	Permissions []string `json:"permissions" swagger:"required"`
}
