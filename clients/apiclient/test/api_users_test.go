/*
Wasp API

Testing UsersApiService

*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech);

package openapi

import (
    "context"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "testing"
    openapiclient "./openapi"
)

func Test_openapi_UsersApiService(t *testing.T) {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)

    t.Run("Test UsersApiService AddUser", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        resp, httpRes, err := apiClient.UsersApi.AddUser(context.Background()).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test UsersApiService ChangeUserPassword", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        var username string

        resp, httpRes, err := apiClient.UsersApi.ChangeUserPassword(context.Background(), username).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test UsersApiService ChangeUserPermissions", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        var username string

        resp, httpRes, err := apiClient.UsersApi.ChangeUserPermissions(context.Background(), username).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test UsersApiService DeleteUser", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        var username string

        resp, httpRes, err := apiClient.UsersApi.DeleteUser(context.Background(), username).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test UsersApiService GetUser", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        var username string

        resp, httpRes, err := apiClient.UsersApi.GetUser(context.Background(), username).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test UsersApiService GetUsers", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        resp, httpRes, err := apiClient.UsersApi.GetUsers(context.Background()).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

}