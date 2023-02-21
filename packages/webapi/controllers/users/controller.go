package users

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

type Controller struct {
	userService interfaces.UserService
}

func NewUsersController(userService interfaces.UserService) interfaces.APIController {
	return &Controller{
		userService: userService,
	}
}

func (c *Controller) Name() string {
	return "users"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("users", c.getUsers, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusOK, "A list of all users", mocker.Get([]models.User{}), nil).
		SetOperationId("getUsers").
		SetSummary("Get a list of all users")

	adminAPI.GET("users/:username", c.getUser, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", params.ParamUsername, params.DescriptionUsername).
		AddResponse(http.StatusNotFound, "User not found", nil, nil).
		AddResponse(http.StatusOK, "Returns a specific user", mocker.Get(models.User{}), nil).
		SetOperationId("getUser").
		SetSummary("Get a user")

	adminAPI.DELETE("users/:username", c.deleteUser, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamUsername, params.DescriptionUsername).
		AddResponse(http.StatusNotFound, "User not found", nil, nil).
		AddResponse(http.StatusOK, "Deletes a specific user", nil, nil).
		SetOperationId("deleteUser").
		SetSummary("Deletes a user")

	adminAPI.POST("users", c.addUser, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamBody(mocker.Get(models.AddUserRequest{}), "", "The user data", true).
		AddResponse(http.StatusBadRequest, "Invalid request", nil, nil).
		AddResponse(http.StatusCreated, "User successfully added", nil, nil).
		SetOperationId("addUser").
		SetSummary("Add a user")

	adminAPI.PUT("users/:username/permissions", c.updateUserPermissions, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamUsername, params.DescriptionUsername).
		AddParamBody(mocker.Get(models.UpdateUserPermissionsRequest{}), "", "The users new permissions", true).
		AddResponse(http.StatusBadRequest, "Invalid request", nil, nil).
		AddResponse(http.StatusNotFound, "User not found", nil, nil).
		AddResponse(http.StatusOK, "User successfully updated", nil, nil).
		SetOperationId("changeUserPermissions").
		SetSummary("Change user permissions")

	adminAPI.PUT("users/:username/password", c.updateUserPassword, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamUsername, params.DescriptionUsername).
		AddParamBody(mocker.Get(models.UpdateUserPasswordRequest{}), "", "The users new password", true).
		AddResponse(http.StatusBadRequest, "Invalid request", nil, nil).
		AddResponse(http.StatusNotFound, "User not found", nil, nil).
		AddResponse(http.StatusOK, "User successfully updated", nil, nil).
		SetOperationId("changeUserPassword").
		SetSummary("Change user password")
}
