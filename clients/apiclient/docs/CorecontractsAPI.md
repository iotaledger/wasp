# \CorecontractsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AccountsGetAccountBalance**](CorecontractsAPI.md#AccountsGetAccountBalance) | **Get** /v1/chain/core/accounts/account/{agentID}/balance | Get all assets belonging to an account
[**AccountsGetAccountFoundries**](CorecontractsAPI.md#AccountsGetAccountFoundries) | **Get** /v1/chain/core/accounts/account/{agentID}/foundries | Get all foundries owned by an account
[**AccountsGetAccountNFTIDs**](CorecontractsAPI.md#AccountsGetAccountNFTIDs) | **Get** /v1/chain/core/accounts/account/{agentID}/nfts | Get all NFT ids belonging to an account
[**AccountsGetAccountNonce**](CorecontractsAPI.md#AccountsGetAccountNonce) | **Get** /v1/chain/core/accounts/account/{agentID}/nonce | Get the current nonce of an account
[**AccountsGetFoundryOutput**](CorecontractsAPI.md#AccountsGetFoundryOutput) | **Get** /v1/chain/core/accounts/foundry_output/{serialNumber} | Get the foundry output
[**AccountsGetNFTData**](CorecontractsAPI.md#AccountsGetNFTData) | **Get** /v1/chain/core/accounts/nftdata/{nftID} | Get the NFT data by an ID
[**AccountsGetNativeTokenIDRegistry**](CorecontractsAPI.md#AccountsGetNativeTokenIDRegistry) | **Get** /v1/chain/core/accounts/token_registry | Get a list of all registries
[**AccountsGetTotalAssets**](CorecontractsAPI.md#AccountsGetTotalAssets) | **Get** /v1/chain/core/accounts/total_assets | Get all stored assets
[**BlocklogGetBlockInfo**](CorecontractsAPI.md#BlocklogGetBlockInfo) | **Get** /v1/chain/core/blocklog/blocks/{blockIndex} | Get the block info of a certain block index
[**BlocklogGetControlAddresses**](CorecontractsAPI.md#BlocklogGetControlAddresses) | **Get** /v1/chain/core/blocklog/controladdresses | Get the control addresses
[**BlocklogGetEventsOfBlock**](CorecontractsAPI.md#BlocklogGetEventsOfBlock) | **Get** /v1/chain/core/blocklog/events/block/{blockIndex} | Get events of a block
[**BlocklogGetEventsOfLatestBlock**](CorecontractsAPI.md#BlocklogGetEventsOfLatestBlock) | **Get** /v1/chain/core/blocklog/events/block/latest | Get events of the latest block
[**BlocklogGetEventsOfRequest**](CorecontractsAPI.md#BlocklogGetEventsOfRequest) | **Get** /v1/chain/core/blocklog/events/request/{requestID} | Get events of a request
[**BlocklogGetLatestBlockInfo**](CorecontractsAPI.md#BlocklogGetLatestBlockInfo) | **Get** /v1/chain/core/blocklog/blocks/latest | Get the block info of the latest block
[**BlocklogGetRequestIDsForBlock**](CorecontractsAPI.md#BlocklogGetRequestIDsForBlock) | **Get** /v1/chain/core/blocklog/blocks/{blockIndex}/requestids | Get the request ids for a certain block index
[**BlocklogGetRequestIDsForLatestBlock**](CorecontractsAPI.md#BlocklogGetRequestIDsForLatestBlock) | **Get** /v1/chain/core/blocklog/blocks/latest/requestids | Get the request ids for the latest block
[**BlocklogGetRequestIsProcessed**](CorecontractsAPI.md#BlocklogGetRequestIsProcessed) | **Get** /v1/chain/core/blocklog/requests/{requestID}/is_processed | Get the request processing status
[**BlocklogGetRequestReceipt**](CorecontractsAPI.md#BlocklogGetRequestReceipt) | **Get** /v1/chain/core/blocklog/requests/{requestID} | Get the receipt of a certain request id
[**BlocklogGetRequestReceiptsOfBlock**](CorecontractsAPI.md#BlocklogGetRequestReceiptsOfBlock) | **Get** /v1/chain/core/blocklog/blocks/{blockIndex}/receipts | Get all receipts of a certain block
[**BlocklogGetRequestReceiptsOfLatestBlock**](CorecontractsAPI.md#BlocklogGetRequestReceiptsOfLatestBlock) | **Get** /v1/chain/core/blocklog/blocks/latest/receipts | Get all receipts of the latest block
[**ErrorsGetErrorMessageFormat**](CorecontractsAPI.md#ErrorsGetErrorMessageFormat) | **Get** /v1/chain/core/errors/{contractHname}/message/{errorID} | Get the error message format of a specific error id
[**GovernanceGetChainAdmin**](CorecontractsAPI.md#GovernanceGetChainAdmin) | **Get** /v1/chain/core/governance/chainadmin | Get the chain admin
[**GovernanceGetChainInfo**](CorecontractsAPI.md#GovernanceGetChainInfo) | **Get** /v1/chain/core/governance/chaininfo | Get the chain info



## AccountsGetAccountBalance

> AssetsResponse AccountsGetAccountBalance(ctx, agentID).Block(block).Execute()

Get all assets belonging to an account

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	agentID := "agentID_example" // string | AgentID (Hex Address for L1 accounts | Hex for EVM)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.AccountsGetAccountBalance(context.Background(), agentID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetAccountBalance``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccountsGetAccountBalance`: AssetsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.AccountsGetAccountBalance`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**agentID** | **string** | AgentID (Hex Address for L1 accounts | Hex for EVM) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetAccountBalanceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**AssetsResponse**](AssetsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetAccountFoundries

> AccountFoundriesResponse AccountsGetAccountFoundries(ctx, chainID, agentID).Block(block).Execute()

Get all foundries owned by an account

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	chainID := "chainID_example" // string | ChainID (Hex Address)
	agentID := "agentID_example" // string | AgentID (Hex Address for L1 accounts, Hex for EVM)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.AccountsGetAccountFoundries(context.Background(), chainID, agentID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetAccountFoundries``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccountsGetAccountFoundries`: AccountFoundriesResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.AccountsGetAccountFoundries`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Hex Address) | 
**agentID** | **string** | AgentID (Hex Address for L1 accounts, Hex for EVM) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetAccountFoundriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **block** | **string** | Block index or trie root | 

### Return type

[**AccountFoundriesResponse**](AccountFoundriesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetAccountNFTIDs

> AccountNFTsResponse AccountsGetAccountNFTIDs(ctx, agentID).Block(block).Execute()

Get all NFT ids belonging to an account

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	agentID := "agentID_example" // string | AgentID (Hex Address for L1 accounts | Hex for EVM)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.AccountsGetAccountNFTIDs(context.Background(), agentID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetAccountNFTIDs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccountsGetAccountNFTIDs`: AccountNFTsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.AccountsGetAccountNFTIDs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**agentID** | **string** | AgentID (Hex Address for L1 accounts | Hex for EVM) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetAccountNFTIDsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**AccountNFTsResponse**](AccountNFTsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetAccountNonce

> AccountNonceResponse AccountsGetAccountNonce(ctx, agentID).Block(block).Execute()

Get the current nonce of an account

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	agentID := "agentID_example" // string | AgentID (Hex Address for L1 accounts | Hex for EVM)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.AccountsGetAccountNonce(context.Background(), agentID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetAccountNonce``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccountsGetAccountNonce`: AccountNonceResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.AccountsGetAccountNonce`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**agentID** | **string** | AgentID (Hex Address for L1 accounts | Hex for EVM) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetAccountNonceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**AccountNonceResponse**](AccountNonceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetFoundryOutput

> FoundryOutputResponse AccountsGetFoundryOutput(ctx, chainID, serialNumber).Block(block).Execute()

Get the foundry output

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	chainID := "chainID_example" // string | ChainID (Hex Address)
	serialNumber := uint32(56) // uint32 | Serial Number (uint32)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.AccountsGetFoundryOutput(context.Background(), chainID, serialNumber).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetFoundryOutput``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccountsGetFoundryOutput`: FoundryOutputResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.AccountsGetFoundryOutput`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Hex Address) | 
**serialNumber** | **uint32** | Serial Number (uint32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetFoundryOutputRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **block** | **string** | Block index or trie root | 

### Return type

[**FoundryOutputResponse**](FoundryOutputResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetNFTData

> AccountsGetNFTData(ctx, nftID).Block(block).Execute()

Get the NFT data by an ID

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	nftID := "nftID_example" // string | NFT ID (Hex)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.CorecontractsAPI.AccountsGetNFTData(context.Background(), nftID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetNFTData``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nftID** | **string** | NFT ID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetNFTDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetNativeTokenIDRegistry

> NativeTokenIDRegistryResponse AccountsGetNativeTokenIDRegistry(ctx).Block(block).Execute()

Get a list of all registries

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.AccountsGetNativeTokenIDRegistry(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetNativeTokenIDRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccountsGetNativeTokenIDRegistry`: NativeTokenIDRegistryResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.AccountsGetNativeTokenIDRegistry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetNativeTokenIDRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**NativeTokenIDRegistryResponse**](NativeTokenIDRegistryResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetTotalAssets

> AssetsResponse AccountsGetTotalAssets(ctx).Block(block).Execute()

Get all stored assets

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.AccountsGetTotalAssets(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.AccountsGetTotalAssets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccountsGetTotalAssets`: AssetsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.AccountsGetTotalAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetTotalAssetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**AssetsResponse**](AssetsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetBlockInfo

> BlockInfoResponse BlocklogGetBlockInfo(ctx, blockIndex).Block(block).Execute()

Get the block info of a certain block index

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetBlockInfo(context.Background(), blockIndex).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetBlockInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetBlockInfo`: BlockInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetBlockInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**blockIndex** | **uint32** | BlockIndex (uint32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetBlockInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**BlockInfoResponse**](BlockInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetControlAddresses

> ControlAddressesResponse BlocklogGetControlAddresses(ctx).Block(block).Execute()

Get the control addresses

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetControlAddresses(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetControlAddresses``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetControlAddresses`: ControlAddressesResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetControlAddresses`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetControlAddressesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**ControlAddressesResponse**](ControlAddressesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetEventsOfBlock

> EventsResponse BlocklogGetEventsOfBlock(ctx, blockIndex).Block(block).Execute()

Get events of a block

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetEventsOfBlock(context.Background(), blockIndex).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetEventsOfBlock``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetEventsOfBlock`: EventsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetEventsOfBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**blockIndex** | **uint32** | BlockIndex (uint32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetEventsOfBlockRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**EventsResponse**](EventsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetEventsOfLatestBlock

> EventsResponse BlocklogGetEventsOfLatestBlock(ctx).Block(block).Execute()

Get events of the latest block

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetEventsOfLatestBlock(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetEventsOfLatestBlock``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetEventsOfLatestBlock`: EventsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetEventsOfLatestBlock`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetEventsOfLatestBlockRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**EventsResponse**](EventsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetEventsOfRequest

> EventsResponse BlocklogGetEventsOfRequest(ctx, requestID).Block(block).Execute()

Get events of a request

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	requestID := "requestID_example" // string | RequestID (Hex)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetEventsOfRequest(context.Background(), requestID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetEventsOfRequest``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetEventsOfRequest`: EventsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetEventsOfRequest`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetEventsOfRequestRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**EventsResponse**](EventsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetLatestBlockInfo

> BlockInfoResponse BlocklogGetLatestBlockInfo(ctx).Block(block).Execute()

Get the block info of the latest block

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetLatestBlockInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetLatestBlockInfo`: BlockInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetLatestBlockInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetLatestBlockInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**BlockInfoResponse**](BlockInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestIDsForBlock

> RequestIDsResponse BlocklogGetRequestIDsForBlock(ctx, blockIndex).Block(block).Execute()

Get the request ids for a certain block index

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetRequestIDsForBlock(context.Background(), blockIndex).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetRequestIDsForBlock``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetRequestIDsForBlock`: RequestIDsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetRequestIDsForBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**blockIndex** | **uint32** | BlockIndex (uint32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestIDsForBlockRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**RequestIDsResponse**](RequestIDsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestIDsForLatestBlock

> RequestIDsResponse BlocklogGetRequestIDsForLatestBlock(ctx).Block(block).Execute()

Get the request ids for the latest block

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetRequestIDsForLatestBlock(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetRequestIDsForLatestBlock``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetRequestIDsForLatestBlock`: RequestIDsResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetRequestIDsForLatestBlock`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestIDsForLatestBlockRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**RequestIDsResponse**](RequestIDsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestIsProcessed

> RequestProcessedResponse BlocklogGetRequestIsProcessed(ctx, requestID).Block(block).Execute()

Get the request processing status

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	requestID := "requestID_example" // string | RequestID (Hex)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetRequestIsProcessed(context.Background(), requestID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetRequestIsProcessed``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetRequestIsProcessed`: RequestProcessedResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetRequestIsProcessed`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestIsProcessedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**RequestProcessedResponse**](RequestProcessedResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestReceipt

> ReceiptResponse BlocklogGetRequestReceipt(ctx, requestID).Block(block).Execute()

Get the receipt of a certain request id

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	requestID := "requestID_example" // string | RequestID (Hex)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetRequestReceipt(context.Background(), requestID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetRequestReceipt``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetRequestReceipt`: ReceiptResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetRequestReceipt`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestReceiptRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestReceiptsOfBlock

> []ReceiptResponse BlocklogGetRequestReceiptsOfBlock(ctx, blockIndex).Block(block).Execute()

Get all receipts of a certain block

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetRequestReceiptsOfBlock(context.Background(), blockIndex).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetRequestReceiptsOfBlock``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetRequestReceiptsOfBlock`: []ReceiptResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetRequestReceiptsOfBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**blockIndex** | **uint32** | BlockIndex (uint32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestReceiptsOfBlockRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**[]ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestReceiptsOfLatestBlock

> []ReceiptResponse BlocklogGetRequestReceiptsOfLatestBlock(ctx).Block(block).Execute()

Get all receipts of the latest block

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.BlocklogGetRequestReceiptsOfLatestBlock(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.BlocklogGetRequestReceiptsOfLatestBlock``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BlocklogGetRequestReceiptsOfLatestBlock`: []ReceiptResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.BlocklogGetRequestReceiptsOfLatestBlock`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestReceiptsOfLatestBlockRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**[]ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ErrorsGetErrorMessageFormat

> ErrorMessageFormatResponse ErrorsGetErrorMessageFormat(ctx, chainID, contractHname, errorID).Block(block).Execute()

Get the error message format of a specific error id

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	chainID := "chainID_example" // string | ChainID (Hex Address)
	contractHname := "contractHname_example" // string | Contract (Hname as Hex)
	errorID := uint32(56) // uint32 | Error Id (uint16)
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.ErrorsGetErrorMessageFormat(context.Background(), chainID, contractHname, errorID).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.ErrorsGetErrorMessageFormat``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ErrorsGetErrorMessageFormat`: ErrorMessageFormatResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.ErrorsGetErrorMessageFormat`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Hex Address) | 
**contractHname** | **string** | Contract (Hname as Hex) | 
**errorID** | **uint32** | Error Id (uint16) | 

### Other Parameters

Other parameters are passed through a pointer to a apiErrorsGetErrorMessageFormatRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **block** | **string** | Block index or trie root | 

### Return type

[**ErrorMessageFormatResponse**](ErrorMessageFormatResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GovernanceGetChainAdmin

> GovChainAdminResponse GovernanceGetChainAdmin(ctx).Block(block).Execute()

Get the chain admin



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.GovernanceGetChainAdmin(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.GovernanceGetChainAdmin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GovernanceGetChainAdmin`: GovChainAdminResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.GovernanceGetChainAdmin`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGovernanceGetChainAdminRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**GovChainAdminResponse**](GovChainAdminResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GovernanceGetChainInfo

> GovChainInfoResponse GovernanceGetChainInfo(ctx).Block(block).Execute()

Get the chain info



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CorecontractsAPI.GovernanceGetChainInfo(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsAPI.GovernanceGetChainInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GovernanceGetChainInfo`: GovChainInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `CorecontractsAPI.GovernanceGetChainInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGovernanceGetChainInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**GovChainInfoResponse**](GovChainInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

