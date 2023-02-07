# \CorecontractsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AccountsGetAccountBalance**](CorecontractsApi.md#AccountsGetAccountBalance) | **Get** /chains/{chainID}/core/accounts/account/{agentID}/balance | Get all assets belonging to an account
[**AccountsGetAccountNFTIDs**](CorecontractsApi.md#AccountsGetAccountNFTIDs) | **Get** /chains/{chainID}/core/accounts/account/{agentID}/nfts | Get all NFT ids belonging to an account
[**AccountsGetAccountNonce**](CorecontractsApi.md#AccountsGetAccountNonce) | **Get** /chains/{chainID}/core/accounts/account/{agentID}/nonce | Get the current nonce of an account
[**AccountsGetAccounts**](CorecontractsApi.md#AccountsGetAccounts) | **Get** /chains/{chainID}/core/accounts | Get a list of all accounts
[**AccountsGetFoundryOutput**](CorecontractsApi.md#AccountsGetFoundryOutput) | **Get** /chains/{chainID}/core/accounts/foundry_output/{serialNumber} | Get the foundry output
[**AccountsGetNFTData**](CorecontractsApi.md#AccountsGetNFTData) | **Get** /chains/{chainID}/core/accounts/nftdata | Get the NFT data by an ID
[**AccountsGetNativeTokenIDRegistry**](CorecontractsApi.md#AccountsGetNativeTokenIDRegistry) | **Get** /chains/{chainID}/core/accounts/token_registry | Get a list of all registries
[**AccountsGetTotalAssets**](CorecontractsApi.md#AccountsGetTotalAssets) | **Get** /chains/{chainID}/core/accounts/total_assets | Get all stored assets
[**BlobsGetAllBlobs**](CorecontractsApi.md#BlobsGetAllBlobs) | **Get** /chains/{chainID}/core/blobs | Get all stored blobs
[**BlobsGetBlobInfo**](CorecontractsApi.md#BlobsGetBlobInfo) | **Get** /chains/{chainID}/core/blobs/{blobHash} | Get all fields of a blob
[**BlobsGetBlobValue**](CorecontractsApi.md#BlobsGetBlobValue) | **Get** /chains/{chainID}/core/blobs/{blobHash}/data/{fieldKey} | Get the value of the supplied field (key)
[**BlocklogGetBlockInfo**](CorecontractsApi.md#BlocklogGetBlockInfo) | **Get** /chains/{chainID}/core/blocklog/blocks/{blockIndex} | Get the block info of a certain block index
[**BlocklogGetControlAddresses**](CorecontractsApi.md#BlocklogGetControlAddresses) | **Get** /chains/{chainID}/core/blocklog/controladdresses | Get the control addresses
[**BlocklogGetEventsOfBlock**](CorecontractsApi.md#BlocklogGetEventsOfBlock) | **Get** /chains/{chainID}/core/blocklog/events/block/{blockIndex} | Get events of a block
[**BlocklogGetEventsOfContract**](CorecontractsApi.md#BlocklogGetEventsOfContract) | **Get** /chains/{chainID}/core/blocklog/events/contract/{contractHname} | Get events of a contract
[**BlocklogGetEventsOfLatestBlock**](CorecontractsApi.md#BlocklogGetEventsOfLatestBlock) | **Get** /chains/{chainID}/core/blocklog/events/block/latest | Get events of the latest block
[**BlocklogGetEventsOfRequest**](CorecontractsApi.md#BlocklogGetEventsOfRequest) | **Get** /chains/{chainID}/core/blocklog/events/request/{requestID} | Get events of a request
[**BlocklogGetLatestBlockInfo**](CorecontractsApi.md#BlocklogGetLatestBlockInfo) | **Get** /chains/{chainID}/core/blocklog/blocks/latest | Get the block info of the latest block
[**BlocklogGetRequestIDsForBlock**](CorecontractsApi.md#BlocklogGetRequestIDsForBlock) | **Get** /chains/{chainID}/core/blocklog/blocks/{blockIndex}/requestids | Get the request ids for a certain block index
[**BlocklogGetRequestIDsForLatestBlock**](CorecontractsApi.md#BlocklogGetRequestIDsForLatestBlock) | **Get** /chains/{chainID}/core/blocklog/blocks/latest/requestids | Get the request ids for the latest block
[**BlocklogGetRequestIsProcessed**](CorecontractsApi.md#BlocklogGetRequestIsProcessed) | **Get** /chains/{chainID}/core/blocklog/requests/{requestID}/is_processed | Get the request processing status
[**BlocklogGetRequestReceipt**](CorecontractsApi.md#BlocklogGetRequestReceipt) | **Get** /chains/{chainID}/core/blocklog/requests/{requestID} | Get the receipt of a certain request id
[**BlocklogGetRequestReceiptsOfBlock**](CorecontractsApi.md#BlocklogGetRequestReceiptsOfBlock) | **Get** /chains/{chainID}/core/blocklog/blocks/{blockIndex}/receipts | Get all receipts of a certain block
[**BlocklogGetRequestReceiptsOfLatestBlock**](CorecontractsApi.md#BlocklogGetRequestReceiptsOfLatestBlock) | **Get** /chains/{chainID}/core/blocklog/blocks/latest/receipts | Get all receipts of the latest block
[**ErrorsGetErrorMessageFormat**](CorecontractsApi.md#ErrorsGetErrorMessageFormat) | **Get** /chains/{chainID}/core/errors/{contractHname}/message/{errorID} | Get the error message format of a specific error id
[**GovernanceGetAllowedStateControllerAddresses**](CorecontractsApi.md#GovernanceGetAllowedStateControllerAddresses) | **Get** /chains/{chainID}/core/governance/allowedstatecontrollers | Get the allowed state controller addresses
[**GovernanceGetChainInfo**](CorecontractsApi.md#GovernanceGetChainInfo) | **Get** /chains/{chainID}/core/governance/chaininfo | Get the chain info
[**GovernanceGetChainOwner**](CorecontractsApi.md#GovernanceGetChainOwner) | **Get** /chains/{chainID}/core/governance/chainowner | Get the chain owner



## AccountsGetAccountBalance

> AssetsResponse AccountsGetAccountBalance(ctx, chainID, agentID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccountBalance(context.Background(), chainID, agentID).Execute()
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



### Return type

[**AssetsResponse**](AssetsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetAccountNFTIDs

> AccountNFTsResponse AccountsGetAccountNFTIDs(ctx, chainID, agentID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccountNFTIDs(context.Background(), chainID, agentID).Execute()
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



### Return type

[**AccountNFTsResponse**](AccountNFTsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetAccountNonce

> AccountNonceResponse AccountsGetAccountNonce(ctx, chainID, agentID).Execute()

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
    agentID := "agentID_example" // string | AgentID (Bech32 for WasmVM | Hex for EVM | '000000@Bech32' Addresses require urlencode)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccountNonce(context.Background(), chainID, agentID).Execute()
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
**agentID** | **string** | AgentID (Bech32 for WasmVM | Hex for EVM | &#39;000000@Bech32&#39; Addresses require urlencode) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetAccountNonceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**AccountNonceResponse**](AccountNonceResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetAccounts

> AccountListResponse AccountsGetAccounts(ctx, chainID).Execute()

Get a list of all accounts

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetAccounts(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetAccounts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetAccounts`: AccountListResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.AccountsGetAccounts`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAccountsGetAccountsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**AccountListResponse**](AccountListResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetFoundryOutput

> FoundryOutputResponse AccountsGetFoundryOutput(ctx, chainID, serialNumber).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetFoundryOutput(context.Background(), chainID, serialNumber).Execute()
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



### Return type

[**FoundryOutputResponse**](FoundryOutputResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetNFTData

> NFTDataResponse AccountsGetNFTData(ctx, chainID, nftID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetNFTData(context.Background(), chainID, nftID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.AccountsGetNFTData``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccountsGetNFTData`: NFTDataResponse
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



### Return type

[**NFTDataResponse**](NFTDataResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetNativeTokenIDRegistry

> NativeTokenIDRegistryResponse AccountsGetNativeTokenIDRegistry(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetNativeTokenIDRegistry(context.Background(), chainID).Execute()
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


### Return type

[**NativeTokenIDRegistryResponse**](NativeTokenIDRegistryResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccountsGetTotalAssets

> AssetsResponse AccountsGetTotalAssets(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.AccountsGetTotalAssets(context.Background(), chainID).Execute()
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


### Return type

[**AssetsResponse**](AssetsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlobsGetAllBlobs

> BlobListResponse BlobsGetAllBlobs(ctx, chainID).Execute()

Get all stored blobs

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlobsGetAllBlobs(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlobsGetAllBlobs``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlobsGetAllBlobs`: BlobListResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlobsGetAllBlobs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlobsGetAllBlobsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**BlobListResponse**](BlobListResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlobsGetBlobInfo

> BlobInfoResponse BlobsGetBlobInfo(ctx, chainID, blobHash).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlobsGetBlobInfo(context.Background(), chainID, blobHash).Execute()
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



### Return type

[**BlobInfoResponse**](BlobInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlobsGetBlobValue

> BlobValueResponse BlobsGetBlobValue(ctx, chainID, blobHash, fieldKey).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlobsGetBlobValue(context.Background(), chainID, blobHash, fieldKey).Execute()
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




### Return type

[**BlobValueResponse**](BlobValueResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetBlockInfo

> BlockInfoResponse BlocklogGetBlockInfo(ctx, chainID, blockIndex).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetBlockInfo(context.Background(), chainID, blockIndex).Execute()
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



### Return type

[**BlockInfoResponse**](BlockInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetControlAddresses

> ControlAddressesResponse BlocklogGetControlAddresses(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetControlAddresses(context.Background(), chainID).Execute()
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


### Return type

[**ControlAddressesResponse**](ControlAddressesResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetEventsOfBlock

> EventsResponse BlocklogGetEventsOfBlock(ctx, chainID, blockIndex).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetEventsOfBlock(context.Background(), chainID, blockIndex).Execute()
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



### Return type

[**EventsResponse**](EventsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetEventsOfContract

> EventsResponse BlocklogGetEventsOfContract(ctx, chainID, contractHname).Execute()

Get events of a contract

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
    contractHname := "contractHname_example" // string | Contract (Hname)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetEventsOfContract(context.Background(), chainID, contractHname).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetEventsOfContract``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetEventsOfContract`: EventsResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetEventsOfContract`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**contractHname** | **string** | Contract (Hname) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetEventsOfContractRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**EventsResponse**](EventsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetEventsOfLatestBlock

> EventsResponse BlocklogGetEventsOfLatestBlock(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetEventsOfLatestBlock(context.Background(), chainID).Execute()
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


### Return type

[**EventsResponse**](EventsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetEventsOfRequest

> EventsResponse BlocklogGetEventsOfRequest(ctx, chainID, requestID).Execute()

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
    requestID := "requestID_example" // string | Request ID (ISCRequestID)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetEventsOfRequest(context.Background(), chainID, requestID).Execute()
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
**requestID** | **string** | Request ID (ISCRequestID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetEventsOfRequestRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**EventsResponse**](EventsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetLatestBlockInfo

> BlockInfoResponse BlocklogGetLatestBlockInfo(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetLatestBlockInfo(context.Background(), chainID).Execute()
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


### Return type

[**BlockInfoResponse**](BlockInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestIDsForBlock

> RequestIDsResponse BlocklogGetRequestIDsForBlock(ctx, chainID, blockIndex).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestIDsForBlock(context.Background(), chainID, blockIndex).Execute()
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



### Return type

[**RequestIDsResponse**](RequestIDsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestIDsForLatestBlock

> RequestIDsResponse BlocklogGetRequestIDsForLatestBlock(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestIDsForLatestBlock(context.Background(), chainID).Execute()
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


### Return type

[**RequestIDsResponse**](RequestIDsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestIsProcessed

> RequestProcessedResponse BlocklogGetRequestIsProcessed(ctx, chainID, requestID).Execute()

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
    requestID := "requestID_example" // string | Request ID (ISCRequestID)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestIsProcessed(context.Background(), chainID, requestID).Execute()
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
**requestID** | **string** | Request ID (ISCRequestID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestIsProcessedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RequestProcessedResponse**](RequestProcessedResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestReceipt

> RequestReceiptResponse BlocklogGetRequestReceipt(ctx, chainID, requestID).Execute()

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
    requestID := "requestID_example" // string | Request ID (ISCRequestID)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestReceipt(context.Background(), chainID, requestID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestReceipt``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestReceipt`: RequestReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `CorecontractsApi.BlocklogGetRequestReceipt`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**requestID** | **string** | Request ID (ISCRequestID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiBlocklogGetRequestReceiptRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RequestReceiptResponse**](RequestReceiptResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestReceiptsOfBlock

> BlockReceiptsResponse BlocklogGetRequestReceiptsOfBlock(ctx, chainID, blockIndex).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestReceiptsOfBlock(context.Background(), chainID, blockIndex).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestReceiptsOfBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestReceiptsOfBlock`: BlockReceiptsResponse
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



### Return type

[**BlockReceiptsResponse**](BlockReceiptsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BlocklogGetRequestReceiptsOfLatestBlock

> BlockReceiptsResponse BlocklogGetRequestReceiptsOfLatestBlock(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BlocklogGetRequestReceiptsOfLatestBlock`: BlockReceiptsResponse
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


### Return type

[**BlockReceiptsResponse**](BlockReceiptsResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ErrorsGetErrorMessageFormat

> ErrorMessageFormatResponse ErrorsGetErrorMessageFormat(ctx, chainID, contractHname, errorID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.ErrorsGetErrorMessageFormat(context.Background(), chainID, contractHname, errorID).Execute()
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




### Return type

[**ErrorMessageFormatResponse**](ErrorMessageFormatResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GovernanceGetAllowedStateControllerAddresses

> GovAllowedStateControllerAddressesResponse GovernanceGetAllowedStateControllerAddresses(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.GovernanceGetAllowedStateControllerAddresses(context.Background(), chainID).Execute()
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


### Return type

[**GovAllowedStateControllerAddressesResponse**](GovAllowedStateControllerAddressesResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GovernanceGetChainInfo

> GovChainInfoResponse GovernanceGetChainInfo(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.GovernanceGetChainInfo(context.Background(), chainID).Execute()
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


### Return type

[**GovChainInfoResponse**](GovChainInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GovernanceGetChainOwner

> GovChainOwnerResponse GovernanceGetChainOwner(ctx, chainID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.CorecontractsApi.GovernanceGetChainOwner(context.Background(), chainID).Execute()
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


### Return type

[**GovChainOwnerResponse**](GovChainOwnerResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

