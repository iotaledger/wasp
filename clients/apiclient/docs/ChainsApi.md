# \ChainsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ActivateChain**](ChainsApi.md#ActivateChain) | **Post** /chains/{chainID}/activate | Activate a chain
[**AddAccessNode**](ChainsApi.md#AddAccessNode) | **Put** /chains/{chainID}/access-node/{peer} | Configure a trusted node to be an access node.
[**ChainsChainIDEvmGet**](ChainsApi.md#ChainsChainIDEvmGet) | **Get** /chains/{chainID}/evm | 
[**DeactivateChain**](ChainsApi.md#DeactivateChain) | **Post** /chains/{chainID}/deactivate | Deactivate a chain
[**GetChainInfo**](ChainsApi.md#GetChainInfo) | **Get** /chains/{chainID} | Get information about a specific chain
[**GetChains**](ChainsApi.md#GetChains) | **Get** /chains | Get a list of all chains
[**GetCommitteeInfo**](ChainsApi.md#GetCommitteeInfo) | **Get** /chains/{chainID}/committee | Get information about the deployed committee
[**GetContracts**](ChainsApi.md#GetContracts) | **Get** /chains/{chainID}/contracts | Get all available chain contracts
[**GetRequestIDFromEVMTransactionID**](ChainsApi.md#GetRequestIDFromEVMTransactionID) | **Get** /chains/{chainID}/evm/tx/{txHash} | Get the ISC request ID for the given Ethereum transaction hash
[**GetStateValue**](ChainsApi.md#GetStateValue) | **Get** /chains/{chainID}/state/{stateKey} | Fetch the raw value associated with the given key in the chain state
[**RemoveAccessNode**](ChainsApi.md#RemoveAccessNode) | **Delete** /chains/{chainID}/access-node/{peer} | Remove an access node.
[**SetChainRecord**](ChainsApi.md#SetChainRecord) | **Post** /chains/{chainID}/chainrecord | Sets the chain record.



## ActivateChain

> ActivateChain(ctx, chainID).Execute()

Activate a chain

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
    resp, r, err := apiClient.ChainsApi.ActivateChain(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.ActivateChain``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiActivateChainRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AddAccessNode

> AddAccessNode(ctx, chainID, peer).Execute()

Configure a trusted node to be an access node.

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
    peer := "peer_example" // string | Name or PubKey (hex) of the trusted peer

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.AddAccessNode(context.Background(), chainID, peer).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.AddAccessNode``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**peer** | **string** | Name or PubKey (hex) of the trusted peer | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddAccessNodeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ChainsChainIDEvmGet

> string ChainsChainIDEvmGet(ctx, chainID).Execute()



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
    resp, r, err := apiClient.ChainsApi.ChainsChainIDEvmGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.ChainsChainIDEvmGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainsChainIDEvmGet`: string
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.ChainsChainIDEvmGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainsChainIDEvmGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeactivateChain

> DeactivateChain(ctx, chainID).Execute()

Deactivate a chain

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
    resp, r, err := apiClient.ChainsApi.DeactivateChain(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.DeactivateChain``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeactivateChainRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetChainInfo

> ChainInfoResponse GetChainInfo(ctx, chainID).Execute()

Get information about a specific chain

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
    resp, r, err := apiClient.ChainsApi.GetChainInfo(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetChainInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetChainInfo`: ChainInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetChainInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ChainInfoResponse**](ChainInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetChains

> []ChainInfoResponse GetChains(ctx).Execute()

Get a list of all chains

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.GetChains(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetChains``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetChains`: []ChainInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetChains`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainsRequest struct via the builder pattern


### Return type

[**[]ChainInfoResponse**](ChainInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCommitteeInfo

> CommitteeInfoResponse GetCommitteeInfo(ctx, chainID).Execute()

Get information about the deployed committee

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
    resp, r, err := apiClient.ChainsApi.GetCommitteeInfo(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetCommitteeInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetCommitteeInfo`: CommitteeInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetCommitteeInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetCommitteeInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**CommitteeInfoResponse**](CommitteeInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetContracts

> []ContractInfoResponse GetContracts(ctx, chainID).Execute()

Get all available chain contracts

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
    resp, r, err := apiClient.ChainsApi.GetContracts(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetContracts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetContracts`: []ContractInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetContracts`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetContractsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]ContractInfoResponse**](ContractInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRequestIDFromEVMTransactionID

> RequestIDResponse GetRequestIDFromEVMTransactionID(ctx, chainID, txHash).Execute()

Get the ISC request ID for the given Ethereum transaction hash

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
    txHash := "txHash_example" // string | Transaction hash (Hex)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.GetRequestIDFromEVMTransactionID(context.Background(), chainID, txHash).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetRequestIDFromEVMTransactionID``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRequestIDFromEVMTransactionID`: RequestIDResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetRequestIDFromEVMTransactionID`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**txHash** | **string** | Transaction hash (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRequestIDFromEVMTransactionIDRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RequestIDResponse**](RequestIDResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetStateValue

> StateResponse GetStateValue(ctx, chainID, stateKey).Execute()

Fetch the raw value associated with the given key in the chain state

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
    stateKey := "stateKey_example" // string | State Key (Hex)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.GetStateValue(context.Background(), chainID, stateKey).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetStateValue``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetStateValue`: StateResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetStateValue`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**stateKey** | **string** | State Key (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetStateValueRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**StateResponse**](StateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveAccessNode

> RemoveAccessNode(ctx, chainID, peer).Execute()

Remove an access node.

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
    peer := "peer_example" // string | Name or PubKey (hex) of the trusted peer

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.RemoveAccessNode(context.Background(), chainID, peer).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.RemoveAccessNode``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**peer** | **string** | Name or PubKey (hex) of the trusted peer | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveAccessNodeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetChainRecord

> SetChainRecord(ctx, chainID).ChainRecord(chainRecord).Execute()

Sets the chain record.

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
    chainRecord := *openapiclient.NewChainRecord([]string{"AccessNodes_example"}, false) // ChainRecord | Chain Record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.SetChainRecord(context.Background(), chainID).ChainRecord(chainRecord).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.SetChainRecord``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetChainRecordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **chainRecord** | [**ChainRecord**](ChainRecord.md) | Chain Record | 

### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

