package users

import (
	"errors"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"

	"github.com/labstack/echo/v4"
)

func (c *Controller) addUser(e echo.Context) error {
	var addUserModel models.AddUserRequest

	if err := e.Bind(&addUserModel); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	err := c.userService.AddUser(addUserModel.Username, addUserModel.Password, addUserModel.Permissions)

	if err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.NoContent(http.StatusCreated)
}

func (c *Controller) updateUserPassword(e echo.Context) error {
	userName := e.Param("userName")

	if userName == "" {
		return apierrors.InvalidPropertyError("userName", errors.New("username is empty"))
	}

	var updateUserPasswordModel models.UpdateUserPasswordRequest

	if err := e.Bind(&updateUserPasswordModel); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	err := c.userService.UpdateUserPassword(userName, updateUserPasswordModel.Password)

	if err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) updateUserPermissions(e echo.Context) error {
	userName := e.Param("userName")

	if userName == "" {
		return apierrors.InvalidPropertyError("userName", errors.New("username is empty"))
	}

	var updateUserPermissionsModel models.UpdateUserPermissionsRequest

	if err := e.Bind(&updateUserPermissionsModel); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	// TODO: Later on, compare the permissions with the permission Wasp actually uses.
	err := c.userService.UpdateUserPermissions(userName, updateUserPermissionsModel.Permissions)

	if err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) deleteUser(e echo.Context) error {
	userName := e.Param("userName")

	if userName == "" {
		return apierrors.InvalidPropertyError("userName", errors.New("username is empty"))
	}

	err := c.userService.DeleteUser(userName)

	if err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) getUser(e echo.Context) error {
	userName := e.Param("userName")

	if userName == "" {
		return apierrors.InvalidPropertyError("userName", errors.New("username is empty"))
	}

	user, err := c.userService.GetUser(userName)

	if err != nil {
		return apierrors.UserNotFoundError(userName)
	}

	return e.JSON(http.StatusOK, user)
}

func (c *Controller) getUsers(e echo.Context) error {
	users := c.userService.GetUsers()

	return e.JSON(http.StatusOK, users)
}
