package users

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	userService interfaces.UserService
}

func NewUsersController(log *loggerpkg.Logger, userService interfaces.UserService) interfaces.APIController {
	return &Controller{
		log:         log,
		userService: userService,
	}
}

func (c *Controller) Name() string {
	return "users"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("users", c.getUsers).
		AddResponse(http.StatusOK, "A list of all users", mocker.Get([]models.User{}), nil).
		SetOperationId("getUsers").
		SetSummary("Get a list of all users")

	adminAPI.GET("users/:username", c.getUser).
		AddParamPath("", "username", "The username").
		AddResponse(http.StatusNotFound, "User not found", nil, nil).
		AddResponse(http.StatusOK, "Returns a specific user", mocker.Get(models.User{}), nil).
		SetOperationId("getUser").
		SetSummary("Get a user")

	adminAPI.POST("users", c.addUser).
		AddParamBody(mocker.Get(models.AddUserRequest{}), "body", "The user data", true).
		AddResponse(http.StatusBadRequest, "Invalid request", nil, nil).
		AddResponse(http.StatusCreated, "User successfully added", nil, nil).
		SetOperationId("addUser").
		SetSummary("Add a user")

	adminAPI.PUT("users/:username/permissions", c.updateUserPermissions).
		AddParamPath("", "username", "The username.").
		AddParamBody(mocker.Get(models.UpdateUserPermissionsRequest{}), "body", "The users new permissions", true).
		AddResponse(http.StatusBadRequest, "Invalid request", nil, nil).
		AddResponse(http.StatusOK, "User successfully updated", nil, nil).
		SetOperationId("changeUserPermissions").
		SetSummary("Change user permissions")

	adminAPI.PUT("users/:username/password", c.updateUserPassword).
		AddParamPath("", "username", "The username.").
		AddParamBody(mocker.Get(models.UpdateUserPasswordRequest{}), "body", "The users new password", true).
		AddResponse(http.StatusBadRequest, "Invalid request", nil, nil).
		AddResponse(http.StatusOK, "User successfully updated", nil, nil).
		SetOperationId("changeUserPassword").
		SetSummary("Change user password")
}
