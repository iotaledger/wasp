package models

type User struct {
	Username    string
	Permissions map[string]struct{}
}
