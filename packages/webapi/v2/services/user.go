package services

import (
	"golang.org/x/exp/maps"

	"github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
)

type UserService struct {
	userManager *users.UserManager
}

func NewUserService(userManager *users.UserManager) interfaces.UserService {
	return &UserService{
		userManager: userManager,
	}
}

func permissionsFromMap(mapPermissions map[string]struct{}) []string {
	permissions := make([]string, len(mapPermissions))

	for i, permission := range maps.Keys(mapPermissions) {
		permissions[i] = permission
	}

	return permissions
}

func permissionsToMap(permissions []string) map[string]struct{} {
	mapPermissions := make(map[string]struct{})

	for _, permission := range permissions {
		mapPermissions[permission] = struct{}{}
	}

	return mapPermissions
}

func (u *UserService) AddUser(username, password string, permissions []string) error {
	passwordHash, passwordSalt, err := users.DerivePasswordKey(password)
	if err != nil {
		return err
	}

	return u.userManager.AddUser(&users.User{
		Name:         username,
		PasswordHash: passwordHash,
		PasswordSalt: passwordSalt,
		Permissions:  permissionsToMap(permissions),
	})
}

func (u *UserService) UpdateUserPassword(username, password string) error {
	passwordHash, passwordSalt, err := users.DerivePasswordKey(password)
	if err != nil {
		return err
	}

	return u.userManager.ChangeUserPassword(username, passwordHash, passwordSalt)
}

func (u *UserService) UpdateUserPermissions(username string, permissions []string) error {
	return u.userManager.ChangeUserPermissions(username, permissionsToMap(permissions))
}

func (u *UserService) DeleteUser(username string) error {
	return u.userManager.RemoveUser(username)
}

func (u *UserService) GetUsers() *[]models.User {
	userList := u.userManager.Users()
	userModels := make([]models.User, len(userList))

	for i, user := range maps.Values(userList) {
		userModels[i] = models.User{
			Username:    user.Name,
			Permissions: permissionsFromMap(user.Permissions),
		}
	}

	return &userModels
}

func (u *UserService) GetUser(username string) (*models.User, error) {
	user, err := u.userManager.User(username)
	if err != nil {
		return nil, err
	}

	return &models.User{
		Username:    user.Name,
		Permissions: permissionsFromMap(user.Permissions),
	}, nil
}
