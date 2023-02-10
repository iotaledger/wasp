package users

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) addUser(e echo.Context) error {
	var addUserModel models.AddUserRequest

	if err := e.Bind(&addUserModel); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	if err := c.userService.AddUser(addUserModel.Username, addUserModel.Password, addUserModel.Permissions); err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.NoContent(http.StatusCreated)
}

func (c *Controller) updateUserPassword(e echo.Context) error {
	userName := e.Param("username")

	if userName == "" {
		return apierrors.InvalidPropertyError("username", errors.New("username is empty"))
	}

	var updateUserPasswordModel models.UpdateUserPasswordRequest

	if err := e.Bind(&updateUserPasswordModel); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	if err := c.userService.UpdateUserPassword(userName, updateUserPasswordModel.Password); err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) updateUserPermissions(e echo.Context) error {
	userName := e.Param("username")
	authContext := e.Get("auth").(*authentication.AuthContext)

	if userName == authContext.Name() {
		return apierrors.InvalidPropertyError("username", errors.New("you can't change your own permissions"))
	}

	if userName == "" {
		return apierrors.InvalidPropertyError("username", errors.New("username is empty"))
	}

	var updateUserPermissionsModel models.UpdateUserPermissionsRequest

	if err := e.Bind(&updateUserPermissionsModel); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	// TODO: Later on, compare the permissions with the permission Wasp actually uses.
	if err := c.userService.UpdateUserPermissions(userName, updateUserPermissionsModel.Permissions); err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) deleteUser(e echo.Context) error {
	userName := e.Param("username")
	authContext := e.Get("auth").(*authentication.AuthContext)

	if userName == authContext.Name() {
		return apierrors.InvalidPropertyError("username", errors.New("you can't remove yourself"))
	}

	if userName == "" {
		return apierrors.InvalidPropertyError("username", errors.New("username is empty"))
	}

	if err := c.userService.DeleteUser(userName); err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) getUser(e echo.Context) error {
	userName := e.Param("username")

	if userName == "" {
		return apierrors.InvalidPropertyError("username", errors.New("username is empty"))
	}

	user, err := c.userService.GetUser(userName)
	if err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.JSON(http.StatusOK, user)
}

func (c *Controller) getUsers(e echo.Context) error {
	return e.JSON(http.StatusOK, c.userService.GetUsers())
}
