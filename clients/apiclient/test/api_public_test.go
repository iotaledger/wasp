/*
Wasp API

Testing PublicApiService

*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech);

package apiclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func Test_apiclient_PublicApiService(t *testing.T) {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)

	t.Run("Test PublicApiService ChainChainIDContractContractHnameCallviewFnameGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var contractHname string
		var fname string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewFnameGet(context.Background(), chainID, contractHname, fname).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDContractContractHnameCallviewFnamePost", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var contractHname string
		var fname string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewFnamePost(context.Background(), chainID, contractHname, fname).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var contractHname string
		var functionHname string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet(context.Background(), chainID, contractHname, functionHname).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var contractHname string
		var functionHname string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost(context.Background(), chainID, contractHname, functionHname).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDEvmReqidTxHashGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var txHash string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDEvmReqidTxHashGet(context.Background(), chainID, txHash).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDRequestPost", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string

		httpRes, err := apiClient.PublicApi.ChainChainIDRequestPost(context.Background(), chainID).Execute()

		require.Nil(t, err)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDRequestReqIDReceiptGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var reqID string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDRequestReqIDReceiptGet(context.Background(), chainID, reqID).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDRequestReqIDWaitGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var reqID string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDRequestReqIDWaitGet(context.Background(), chainID, reqID).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDStateKeyGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string
		var key string

		resp, httpRes, err := apiClient.PublicApi.ChainChainIDStateKeyGet(context.Background(), chainID, key).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService ChainChainIDWsGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		var chainID string

		httpRes, err := apiClient.PublicApi.ChainChainIDWsGet(context.Background(), chainID).Execute()

		require.Nil(t, err)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

	t.Run("Test PublicApiService InfoGet", func(t *testing.T) {

		t.Skip("skip test")  // remove to run test

		resp, httpRes, err := apiClient.PublicApi.InfoGet(context.Background()).Execute()

		require.Nil(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, httpRes.StatusCode)

	})

}
