# \CorecontractsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AccountsGetAccountBalance**](CorecontractsApi.md#AccountsGetAccountBalance) | **Get** /v1/chains/{chainID}/core/accounts/account/{agentID}/balance | Get all assets belonging to an account
[**AccountsGetAccountFoundries**](CorecontractsApi.md#AccountsGetAccountFoundries) | **Get** /v1/chains/{chainID}/core/accounts/account/{agentID}/foundries | Get all foundries owned by an account
[**AccountsGetAccountNFTIDs**](CorecontractsApi.md#AccountsGetAccountNFTIDs) | **Get** /v1/chains/{chainID}/core/accounts/account/{agentID}/nfts | Get all NFT ids belonging to an account
[**AccountsGetAccountNonce**](CorecontractsApi.md#AccountsGetAccountNonce) | **Get** /v1/chains/{chainID}/core/accounts/account/{agentID}/nonce | Get the current nonce of an account
[**AccountsGetFoundryOutput**](CorecontractsApi.md#AccountsGetFoundryOutput) | **Get** /v1/chains/{chainID}/core/accounts/foundry_output/{serialNumber} | Get the foundry output
[**AccountsGetNFTData**](CorecontractsApi.md#AccountsGetNFTData) | **Get** /v1/chains/{chainID}/core/accounts/nftdata/{nftID} | Get the NFT data by an ID
[**AccountsGetNativeTokenIDRegistry**](CorecontractsApi.md#AccountsGetNativeTokenIDRegistry) | **Get** /v1/chains/{chainID}/core/accounts/token_registry | Get a list of all registries
[**AccountsGetTotalAssets**](CorecontractsApi.md#AccountsGetTotalAssets) | **Get** /v1/chains/{chainID}/core/accounts/total_assets | Get all stored assets
[**BlobsGetBlobInfo**](CorecontractsApi.md#BlobsGetBlobInfo) | **Get** /v1/chains/{chainID}/core/blobs/{blobHash} | Get all fields of a blob
[**BlobsGetBlobValue**](CorecontractsApi.md#BlobsGetBlobValue) | **Get** /v1/chains/{chainID}/core/blobs/{blobHash}/data/{fieldKey} | Get the value of the supplied field (key)
[**BlocklogGetBlockInfo**](CorecontractsApi.md#BlocklogGetBlockInfo) | **Get** /v1/chains/{chainID}/core/blocklog/blocks/{blockIndex} | Get the block info of a certain block index
[**BlocklogGetControlAddresses**](CorecontractsApi.md#BlocklogGetControlAddresses) | **Get** /v1/chains/{chainID}/core/blocklog/controladdresses | Get the control addresses
[**BlocklogGetEventsOfBlock**](CorecontractsApi.md#BlocklogGetEventsOfBlock) | **Get** /v1/chains/{chainID}/core/blocklog/events/block/{blockIndex} | Get events of a block
[**BlocklogGetEventsOfLatestBlock**](CorecontractsApi.md#BlocklogGetEventsOfLatestBlock) | **Get** /v1/chains/{chainID}/core/blocklog/events/block/latest | Get events of the latest block
[**BlocklogGetEventsOfRequest**](CorecontractsApi.md#BlocklogGetEventsOfRequest) | **Get** /v1/chains/{chainID}/core/blocklog/events/request/{requestID} | Get events of a request
[**BlocklogGetLatestBlockInfo**](CorecontractsApi.md#BlocklogGetLatestBlockInfo) | **Get** /v1/chains/{chainID}/core/blocklog/blocks/latest | Get the block info of the latest block
[**BlocklogGetRequestIDsForBlock**](CorecontractsApi.md#BlocklogGetRequestIDsForBlock) | **Get** /v1/chains/{chainID}/core/blocklog/blocks/{blockIndex}/requestids | Get the request ids for a certain block index
[**BlocklogGetRequestIDsForLatestBlock**](CorecontractsApi.md#BlocklogGetRequestIDsForLatestBlock) | **Get** /v1/chains/{chainID}/core/blocklog/blocks/latest/requestids | Get the request ids for the latest block
[**BlocklogGetRequestIsProcessed**](CorecontractsApi.md#BlocklogGetRequestIsProcessed) | **Get** /v1/chains/{chainID}/core/blocklog/requests/{requestID}/is_processed | Get the request processing status
[**BlocklogGetRequestReceipt**](CorecontractsApi.md#BlocklogGetRequestReceipt) | **Get** /v1/chains/{chainID}/core/blocklog/requests/{requestID} | Get the receipt of a certain request id
[**BlocklogGetRequestReceiptsOfBlock**](CorecontractsApi.md#BlocklogGetRequestReceiptsOfBlock) | **Get** /v1/chains/{chainID}/core/blocklog/blocks/{blockIndex}/receipts | Get all receipts of a certain block
[**BlocklogGetRequestReceiptsOfLatestBlock**](CorecontractsApi.md#BlocklogGetRequestReceiptsOfLatestBlock) | **Get** /v1/chains/{chainID}/core/blocklog/blocks/latest/receipts | Get all receipts of the latest block
[**ErrorsGetErrorMessageFormat**](CorecontractsApi.md#ErrorsGetErrorMessageFormat) | **Get** /v1/chains/{chainID}/core/errors/{contractHname}/message/{errorID} | Get the error message format of a specific error id
[**GovernanceGetAllowedStateControllerAddresses**](CorecontractsApi.md#GovernanceGetAllowedStateControllerAddresses) | **Get** /v1/chains/{chainID}/core/governance/allowedstatecontrollers | Get the allowed state controller addresses
[**GovernanceGetChainInfo**](CorecontractsApi.md#GovernanceGetChainInfo) | **Get** /v1/chains/{chainID}/core/governance/chaininfo | Get the chain info
[**GovernanceGetChainOwner**](CorecontractsApi.md#GovernanceGetChainOwner) | **Get** /v1/chains/{chainID}/core/governance/chainowner | Get the chain owner



## AccountsGetAccountBalance

> AssetsResponse AccountsGetAccountBalance(ctx, chainID, agentID).Block(block).Execute()

Get all assets belonging to an account

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    agentID := "agentID_example" // string | AgentID (Bech32 for WasmVM | Hex for EVM)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccountBalance(context.Background(), chainID, agentID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetAccountBalance``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetAccountBalance`: AssetsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetAccountBalance`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**agentID** | **string** | AgentID (Bech32 for WasmVM | Hex for EVM) | 

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
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    agentID := "agentID_example" // string | AgentID (Bech32 for WasmVM | Hex for EVM)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccountFoundries(context.Background(), chainID, agentID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetAccountFoundries``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetAccountFoundries`: AccountFoundriesResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetAccountFoundries`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**agentID** | **string** | AgentID (Bech32 for WasmVM | Hex for EVM) | 

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

> AccountNFTsResponse AccountsGetAccountNFTIDs(ctx, chainID, agentID).Block(block).Execute()

Get all NFT ids belonging to an account

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    agentID := "agentID_example" // string | AgentID (Bech32 for WasmVM | Hex for EVM)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccountNFTIDs(context.Background(), chainID, agentID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetAccountNFTIDs``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetAccountNFTIDs`: AccountNFTsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetAccountNFTIDs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**agentID** | **string** | AgentID (Bech32 for WasmVM | Hex for EVM) | 

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

> AccountNonceResponse AccountsGetAccountNonce(ctx, chainID, agentID).Block(block).Execute()

Get the current nonce of an account

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    agentID := "agentID_example" // string | AgentID (Bech32 for WasmVM | Hex for EVM)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccountNonce(context.Background(), chainID, agentID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetAccountNonce``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetAccountNonce`: AccountNonceResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetAccountNonce`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**agentID** | **string** | AgentID (Bech32 for WasmVM | Hex for EVM) | 

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
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    serialNumber := uint32(56) // uint32 | Serial Number (uint32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetFoundryOutput(context.Background(), chainID, serialNumber).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetFoundryOutput``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetFoundryOutput`: FoundryOutputResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetFoundryOutput`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> NFTJSON AccountsGetNFTData(ctx, chainID, nftID).Block(block).Execute()

Get the NFT data by an ID

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    nftID := "nftID_example" // string | NFT ID (Hex)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetNFTData(context.Background(), chainID, nftID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetNFTData``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetNFTData`: NFTJSON
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetNFTData`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**nftID** | **string** | NFT ID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetNFTDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **block** | **string** | Block index or trie root | 

### Return type

[**NFTJSON**](NFTJSON.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetNativeTokenIDRegistry

> NativeTokenIDRegistryResponse AccountsGetNativeTokenIDRegistry(ctx, chainID).Block(block).Execute()

Get a list of all registries

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetNativeTokenIDRegistry(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetNativeTokenIDRegistry``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetNativeTokenIDRegistry`: NativeTokenIDRegistryResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetNativeTokenIDRegistry`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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

> AssetsResponse AccountsGetTotalAssets(ctx, chainID).Block(block).Execute()

Get all stored assets

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetTotalAssets(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetTotalAssets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetTotalAssets`: AssetsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetTotalAssets`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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


## BlobsGetBlobInfo

> BlobInfoResponse BlobsGetBlobInfo(ctx, chainID, blobHash).Block(block).Execute()

Get all fields of a blob

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    blobHash := "blobHash_example" // string | BlobHash (Hex)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlobsGetBlobInfo(context.Background(), chainID, blobHash).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlobsGetBlobInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlobsGetBlobInfo`: BlobInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlobsGetBlobInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**blobHash** | **string** | BlobHash (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlobsGetBlobInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **block** | **string** | Block index or trie root | 

### Return type

[**BlobInfoResponse**](BlobInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlobsGetBlobValue

> BlobValueResponse BlobsGetBlobValue(ctx, chainID, blobHash, fieldKey).Block(block).Execute()

Get the value of the supplied field (key)

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    blobHash := "blobHash_example" // string | BlobHash (Hex)
    fieldKey := "fieldKey_example" // string | FieldKey (String)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlobsGetBlobValue(context.Background(), chainID, blobHash, fieldKey).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlobsGetBlobValue``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlobsGetBlobValue`: BlobValueResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlobsGetBlobValue`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**blobHash** | **string** | BlobHash (Hex) | 
**fieldKey** | **string** | FieldKey (String) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlobsGetBlobValueRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **block** | **string** | Block index or trie root | 

### Return type

[**BlobValueResponse**](BlobValueResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetBlockInfo

> BlockInfoResponse BlocklogGetBlockInfo(ctx, chainID, blockIndex).Block(block).Execute()

Get the block info of a certain block index

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetBlockInfo(context.Background(), chainID, blockIndex).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetBlockInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetBlockInfo`: BlockInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetBlockInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> ControlAddressesResponse BlocklogGetControlAddresses(ctx, chainID).Block(block).Execute()

Get the control addresses

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetControlAddresses(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetControlAddresses``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetControlAddresses`: ControlAddressesResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetControlAddresses`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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

> EventsResponse BlocklogGetEventsOfBlock(ctx, chainID, blockIndex).Block(block).Execute()

Get events of a block

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetEventsOfBlock(context.Background(), chainID, blockIndex).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetEventsOfBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetEventsOfBlock`: EventsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetEventsOfBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> EventsResponse BlocklogGetEventsOfLatestBlock(ctx, chainID).Block(block).Execute()

Get events of the latest block

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetEventsOfLatestBlock(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetEventsOfLatestBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetEventsOfLatestBlock`: EventsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetEventsOfLatestBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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

> EventsResponse BlocklogGetEventsOfRequest(ctx, chainID, requestID).Block(block).Execute()

Get events of a request

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    requestID := "requestID_example" // string | RequestID (Hex)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetEventsOfRequest(context.Background(), chainID, requestID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetEventsOfRequest``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetEventsOfRequest`: EventsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetEventsOfRequest`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> BlockInfoResponse BlocklogGetLatestBlockInfo(ctx, chainID).Block(block).Execute()

Get the block info of the latest block

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetLatestBlockInfo(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetLatestBlockInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetLatestBlockInfo`: BlockInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetLatestBlockInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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

> RequestIDsResponse BlocklogGetRequestIDsForBlock(ctx, chainID, blockIndex).Block(block).Execute()

Get the request ids for a certain block index

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestIDsForBlock(context.Background(), chainID, blockIndex).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestIDsForBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestIDsForBlock`: RequestIDsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetRequestIDsForBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> RequestIDsResponse BlocklogGetRequestIDsForLatestBlock(ctx, chainID).Block(block).Execute()

Get the request ids for the latest block

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestIDsForLatestBlock(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestIDsForLatestBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestIDsForLatestBlock`: RequestIDsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetRequestIDsForLatestBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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

> RequestProcessedResponse BlocklogGetRequestIsProcessed(ctx, chainID, requestID).Block(block).Execute()

Get the request processing status

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    requestID := "requestID_example" // string | RequestID (Hex)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestIsProcessed(context.Background(), chainID, requestID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestIsProcessed``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestIsProcessed`: RequestProcessedResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetRequestIsProcessed`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> ReceiptResponse BlocklogGetRequestReceipt(ctx, chainID, requestID).Block(block).Execute()

Get the receipt of a certain request id

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    requestID := "requestID_example" // string | RequestID (Hex)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestReceipt(context.Background(), chainID, requestID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestReceipt``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestReceipt`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetRequestReceipt`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> []ReceiptResponse BlocklogGetRequestReceiptsOfBlock(ctx, chainID, blockIndex).Block(block).Execute()

Get all receipts of a certain block

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    blockIndex := uint32(56) // uint32 | BlockIndex (uint32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestReceiptsOfBlock(context.Background(), chainID, blockIndex).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestReceiptsOfBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestReceiptsOfBlock`: []ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetRequestReceiptsOfBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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

> []ReceiptResponse BlocklogGetRequestReceiptsOfLatestBlock(ctx, chainID).Block(block).Execute()

Get all receipts of the latest block

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestReceiptsOfLatestBlock`: []ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    contractHname := "contractHname_example" // string | Contract (Hname as Hex)
    errorID := uint32(56) // uint32 | Error Id (uint16)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.ErrorsGetErrorMessageFormat(context.Background(), chainID, contractHname, errorID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.ErrorsGetErrorMessageFormat``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ErrorsGetErrorMessageFormat`: ErrorMessageFormatResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.ErrorsGetErrorMessageFormat`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
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


## GovernanceGetAllowedStateControllerAddresses

> GovAllowedStateControllerAddressesResponse GovernanceGetAllowedStateControllerAddresses(ctx, chainID).Block(block).Execute()

Get the allowed state controller addresses



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.GovernanceGetAllowedStateControllerAddresses(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.GovernanceGetAllowedStateControllerAddresses``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GovernanceGetAllowedStateControllerAddresses`: GovAllowedStateControllerAddressesResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.GovernanceGetAllowedStateControllerAddresses`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGovernanceGetAllowedStateControllerAddressesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**GovAllowedStateControllerAddressesResponse**](GovAllowedStateControllerAddressesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GovernanceGetChainInfo

> GovChainInfoResponse GovernanceGetChainInfo(ctx, chainID).Block(block).Execute()

Get the chain info



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.GovernanceGetChainInfo(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.GovernanceGetChainInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GovernanceGetChainInfo`: GovChainInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.GovernanceGetChainInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

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


## GovernanceGetChainOwner

> GovChainOwnerResponse GovernanceGetChainOwner(ctx, chainID).Block(block).Execute()

Get the chain owner



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    chainID := "chainID_example" // string | ChainID (Bech32)
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.GovernanceGetChainOwner(context.Background(), chainID).Block(block).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.GovernanceGetChainOwner``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GovernanceGetChainOwner`: GovChainOwnerResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.GovernanceGetChainOwner`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGovernanceGetChainOwnerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **block** | **string** | Block index or trie root | 

### Return type

[**GovChainOwnerResponse**](GovChainOwnerResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

