/*
Wasp API

Testing MetricsApiService

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

func Test_openapi_MetricsApiService(t *testing.T) {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)

    t.Run("Test MetricsApiService GetChainMetrics", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        var chainID string

        resp, httpRes, err := apiClient.MetricsApi.GetChainMetrics(context.Background(), chainID).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test MetricsApiService GetChainPipeMetrics", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        var chainID string

        resp, httpRes, err := apiClient.MetricsApi.GetChainPipeMetrics(context.Background(), chainID).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test MetricsApiService GetChainWorkflowMetrics", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        var chainID string

        resp, httpRes, err := apiClient.MetricsApi.GetChainWorkflowMetrics(context.Background(), chainID).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

    t.Run("Test MetricsApiService GetL1Metrics", func(t *testing.T) {

        t.Skip("skip test")  // remove to run test

        resp, httpRes, err := apiClient.MetricsApi.GetL1Metrics(context.Background()).Execute()

        require.Nil(t, err)
        require.NotNil(t, resp)
        assert.Equal(t, 200, httpRes.StatusCode)

    })

}