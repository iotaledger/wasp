# \ChainsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ActivateChain**](ChainsApi.md#ActivateChain) | **Post** /v1/chains/{chainID}/activate | Activate a chain
[**AddAccessNode**](ChainsApi.md#AddAccessNode) | **Put** /v1/chains/{chainID}/access-node/{peer} | Configure a trusted node to be an access node.
[**CallView**](ChainsApi.md#CallView) | **Post** /v1/chains/{chainID}/callview | Call a view function on a contract by Hname
[**DeactivateChain**](ChainsApi.md#DeactivateChain) | **Post** /v1/chains/{chainID}/deactivate | Deactivate a chain
[**DumpAccounts**](ChainsApi.md#DumpAccounts) | **Post** /v1/chains/{chainID}/dump-accounts | dump accounts information into a humanly-readable format
[**EstimateGasOffledger**](ChainsApi.md#EstimateGasOffledger) | **Post** /v1/chains/{chainID}/estimategas-offledger | Estimates gas for a given off-ledger ISC request
[**EstimateGasOnledger**](ChainsApi.md#EstimateGasOnledger) | **Post** /v1/chains/{chainID}/estimategas-onledger | Estimates gas for a given on-ledger ISC request
[**GetChainInfo**](ChainsApi.md#GetChainInfo) | **Get** /v1/chains/{chainID} | Get information about a specific chain
[**GetChains**](ChainsApi.md#GetChains) | **Get** /v1/chains | Get a list of all chains
[**GetCommitteeInfo**](ChainsApi.md#GetCommitteeInfo) | **Get** /v1/chains/{chainID}/committee | Get information about the deployed committee
[**GetContracts**](ChainsApi.md#GetContracts) | **Get** /v1/chains/{chainID}/contracts | Get all available chain contracts
[**GetMempoolContents**](ChainsApi.md#GetMempoolContents) | **Get** /v1/chains/{chainID}/mempool | Get the contents of the mempool.
[**GetReceipt**](ChainsApi.md#GetReceipt) | **Get** /v1/chains/{chainID}/receipts/{requestID} | Get a receipt from a request ID
[**GetStateValue**](ChainsApi.md#GetStateValue) | **Get** /v1/chains/{chainID}/state/{stateKey} | Fetch the raw value associated with the given key in the chain state
[**RemoveAccessNode**](ChainsApi.md#RemoveAccessNode) | **Delete** /v1/chains/{chainID}/access-node/{peer} | Remove an access node.
[**SetChainRecord**](ChainsApi.md#SetChainRecord) | **Post** /v1/chains/{chainID}/chainrecord | Sets the chain record.
[**V1ChainsChainIDEvmPost**](ChainsApi.md#V1ChainsChainIDEvmPost) | **Post** /v1/chains/{chainID}/evm | Ethereum JSON-RPC
[**V1ChainsChainIDEvmWsGet**](ChainsApi.md#V1ChainsChainIDEvmWsGet) | **Get** /v1/chains/{chainID}/evm/ws | Ethereum JSON-RPC (Websocket transport)
[**WaitForRequest**](ChainsApi.md#WaitForRequest) | **Get** /v1/chains/{chainID}/requests/{requestID}/wait | Wait until the given request has been processed by the node



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


## CallView

> JSONDict CallView(ctx, chainID).ContractCallViewRequest(contractCallViewRequest).Execute()

Call a view function on a contract by Hname



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
    contractCallViewRequest := *openapiclient.NewContractCallViewRequest(*openapiclient.NewJSONDict(), "ContractHName_example", "ContractName_example", "FunctionHName_example", "FunctionName_example") // ContractCallViewRequest | Parameters

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.CallView(context.Background(), chainID).ContractCallViewRequest(contractCallViewRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.CallView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CallView`: JSONDict
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.CallView`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiCallViewRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **contractCallViewRequest** | [**ContractCallViewRequest**](ContractCallViewRequest.md) | Parameters | 

### Return type

[**JSONDict**](JSONDict.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
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


## DumpAccounts

> DumpAccounts(ctx, chainID).Execute()

dump accounts information into a humanly-readable format

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
    resp, r, err := apiClient.ChainsApi.DumpAccounts(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.DumpAccounts``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDumpAccountsRequest struct via the builder pattern


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


## EstimateGasOffledger

> ReceiptResponse EstimateGasOffledger(ctx, chainID).Request(request).Execute()

Estimates gas for a given off-ledger ISC request

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
    request := *openapiclient.NewEstimateGasRequestOffledger("FromAddress_example", "RequestBytes_example") // EstimateGasRequestOffledger | Request

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.EstimateGasOffledger(context.Background(), chainID).Request(request).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.EstimateGasOffledger``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `EstimateGasOffledger`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.EstimateGasOffledger`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiEstimateGasOffledgerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**EstimateGasRequestOffledger**](EstimateGasRequestOffledger.md) | Request | 

### Return type

[**ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## EstimateGasOnledger

> ReceiptResponse EstimateGasOnledger(ctx, chainID).Request(request).Execute()

Estimates gas for a given on-ledger ISC request

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
    request := *openapiclient.NewEstimateGasRequestOnledger("OutputBytes_example") // EstimateGasRequestOnledger | Request

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.EstimateGasOnledger(context.Background(), chainID).Request(request).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.EstimateGasOnledger``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `EstimateGasOnledger`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.EstimateGasOnledger`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiEstimateGasOnledgerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**EstimateGasRequestOnledger**](EstimateGasRequestOnledger.md) | Request | 

### Return type

[**ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetChainInfo

> ChainInfoResponse GetChainInfo(ctx, chainID).Block(block).Execute()

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
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.GetChainInfo(context.Background(), chainID).Block(block).Execute()
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

 **block** | **string** | Block index or trie root | 

### Return type

[**ChainInfoResponse**](ChainInfoResponse.md)

### Authorization

No authorization required

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

> CommitteeInfoResponse GetCommitteeInfo(ctx, chainID).Block(block).Execute()

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
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.GetCommitteeInfo(context.Background(), chainID).Block(block).Execute()
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

 **block** | **string** | Block index or trie root | 

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

> []ContractInfoResponse GetContracts(ctx, chainID).Block(block).Execute()

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
    block := "block_example" // string | Block index or trie root (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.GetContracts(context.Background(), chainID).Block(block).Execute()
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

 **block** | **string** | Block index or trie root | 

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


## GetMempoolContents

> []int32 GetMempoolContents(ctx, chainID).Execute()

Get the contents of the mempool.

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
    resp, r, err := apiClient.ChainsApi.GetMempoolContents(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetMempoolContents``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetMempoolContents`: []int32
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetMempoolContents`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetMempoolContentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**[]int32**

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetReceipt

> ReceiptResponse GetReceipt(ctx, chainID, requestID).Execute()

Get a receipt from a request ID

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.GetReceipt(context.Background(), chainID, requestID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.GetReceipt``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetReceipt`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.GetReceipt`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetReceiptRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



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


## V1ChainsChainIDEvmPost

> V1ChainsChainIDEvmPost(ctx, chainID).Execute()

Ethereum JSON-RPC

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
    resp, r, err := apiClient.ChainsApi.V1ChainsChainIDEvmPost(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.V1ChainsChainIDEvmPost``: %v\n", err)
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

Other parameters are passed through a pointer to a apiV1ChainsChainIDEvmPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1ChainsChainIDEvmWsGet

> V1ChainsChainIDEvmWsGet(ctx, chainID).Execute()

Ethereum JSON-RPC (Websocket transport)

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
    resp, r, err := apiClient.ChainsApi.V1ChainsChainIDEvmWsGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.V1ChainsChainIDEvmWsGet``: %v\n", err)
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

Other parameters are passed through a pointer to a apiV1ChainsChainIDEvmWsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WaitForRequest

> ReceiptResponse WaitForRequest(ctx, chainID, requestID).TimeoutSeconds(timeoutSeconds).WaitForL1Confirmation(waitForL1Confirmation).Execute()

Wait until the given request has been processed by the node

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
    timeoutSeconds := int32(56) // int32 | The timeout in seconds, maximum 60s (optional)
    waitForL1Confirmation := true // bool | Wait for the block to be confirmed on L1 (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ChainsApi.WaitForRequest(context.Background(), chainID, requestID).TimeoutSeconds(timeoutSeconds).WaitForL1Confirmation(waitForL1Confirmation).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ChainsApi.WaitForRequest``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `WaitForRequest`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `ChainsApi.WaitForRequest`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiWaitForRequestRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **timeoutSeconds** | **int32** | The timeout in seconds, maximum 60s | 
 **waitForL1Confirmation** | **bool** | Wait for the block to be confirmed on L1 | 

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

