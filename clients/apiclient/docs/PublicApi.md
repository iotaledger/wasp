# \PublicApi

All URIs are relative to *http://localhost:9090*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ChainChainIDContractContractHnameCallviewFnameGet**](PublicApi.md#ChainChainIDContractContractHnameCallviewFnameGet) | **Get** /chain/{chainID}/contract/{contractHname}/callview/{fname} | Call a view function on a contract by name
[**ChainChainIDContractContractHnameCallviewFnamePost**](PublicApi.md#ChainChainIDContractContractHnameCallviewFnamePost) | **Post** /chain/{chainID}/contract/{contractHname}/callview/{fname} | Call a view function on a contract by name
[**ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet**](PublicApi.md#ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet) | **Get** /chain/{chainID}/contract/{contractHname}/callviewbyhname/{functionHname} | Call a view function on a contract by Hname
[**ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost**](PublicApi.md#ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost) | **Post** /chain/{chainID}/contract/{contractHname}/callviewbyhname/{functionHname} | Call a view function on a contract by Hname
[**ChainChainIDEvmReqidTxHashGet**](PublicApi.md#ChainChainIDEvmReqidTxHashGet) | **Get** /chain/{chainID}/evm/reqid/{txHash} | Get the ISC request ID for the given Ethereum transaction hash
[**ChainChainIDRequestPost**](PublicApi.md#ChainChainIDRequestPost) | **Post** /chain/{chainID}/request | Post an off-ledger request
[**ChainChainIDRequestReqIDReceiptGet**](PublicApi.md#ChainChainIDRequestReqIDReceiptGet) | **Get** /chain/{chainID}/request/{reqID}/receipt | Get the processing status of a given request in the node
[**ChainChainIDRequestReqIDWaitGet**](PublicApi.md#ChainChainIDRequestReqIDWaitGet) | **Get** /chain/{chainID}/request/{reqID}/wait | Wait until the given request has been processed by the node
[**ChainChainIDStateKeyGet**](PublicApi.md#ChainChainIDStateKeyGet) | **Get** /chain/{chainID}/state/{key} | Fetch the raw value associated with the given key in the chain state
[**ChainChainIDWsGet**](PublicApi.md#ChainChainIDWsGet) | **Get** /chain/{chainID}/ws | 
[**InfoGet**](PublicApi.md#InfoGet) | **Get** /info | Get information about the node



## ChainChainIDContractContractHnameCallviewFnameGet

> JSONDict ChainChainIDContractContractHnameCallviewFnameGet(ctx, chainID, contractHname, fname).Params(params).Execute()

Call a view function on a contract by name

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
    chainID := "chainID_example" // string | ChainID
    contractHname := "contractHname_example" // string | Contract Hname
    fname := "fname_example" // string | Function name
    params := *openapiclient.NewJSONDict() // JSONDict | Parameters (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewFnameGet(context.Background(), chainID, contractHname, fname).Params(params).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDContractContractHnameCallviewFnameGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDContractContractHnameCallviewFnameGet`: JSONDict
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDContractContractHnameCallviewFnameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID | 
**contractHname** | **string** | Contract Hname | 
**fname** | **string** | Function name | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDContractContractHnameCallviewFnameGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **params** | [**JSONDict**](JSONDict.md) | Parameters | 

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


## ChainChainIDContractContractHnameCallviewFnamePost

> JSONDict ChainChainIDContractContractHnameCallviewFnamePost(ctx, chainID, contractHname, fname).Params(params).Execute()

Call a view function on a contract by name

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
    chainID := "chainID_example" // string | ChainID
    contractHname := "contractHname_example" // string | Contract Hname
    fname := "fname_example" // string | Function name
    params := *openapiclient.NewJSONDict() // JSONDict | Parameters (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewFnamePost(context.Background(), chainID, contractHname, fname).Params(params).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDContractContractHnameCallviewFnamePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDContractContractHnameCallviewFnamePost`: JSONDict
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDContractContractHnameCallviewFnamePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID | 
**contractHname** | **string** | Contract Hname | 
**fname** | **string** | Function name | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDContractContractHnameCallviewFnamePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **params** | [**JSONDict**](JSONDict.md) | Parameters | 

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


## ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet

> JSONDict ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet(ctx, chainID, contractHname, functionHname).Params(params).Execute()

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
    chainID := "chainID_example" // string | ChainID
    contractHname := "contractHname_example" // string | Contract Hname
    functionHname := "functionHname_example" // string | Function Hname
    params := *openapiclient.NewJSONDict() // JSONDict | Parameters (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet(context.Background(), chainID, contractHname, functionHname).Params(params).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet`: JSONDict
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID | 
**contractHname** | **string** | Contract Hname | 
**functionHname** | **string** | Function Hname | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDContractContractHnameCallviewbyhnameFunctionHnameGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **params** | [**JSONDict**](JSONDict.md) | Parameters | 

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


## ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost

> JSONDict ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost(ctx, chainID, contractHname, functionHname).Params(params).Execute()

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
    chainID := "chainID_example" // string | ChainID
    contractHname := "contractHname_example" // string | Contract Hname
    functionHname := "functionHname_example" // string | Function Hname
    params := *openapiclient.NewJSONDict() // JSONDict | Parameters (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost(context.Background(), chainID, contractHname, functionHname).Params(params).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost`: JSONDict
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID | 
**contractHname** | **string** | Contract Hname | 
**functionHname** | **string** | Function Hname | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDContractContractHnameCallviewbyhnameFunctionHnamePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **params** | [**JSONDict**](JSONDict.md) | Parameters | 

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


## ChainChainIDEvmReqidTxHashGet

> string ChainChainIDEvmReqidTxHashGet(ctx, chainID, txHash).Execute()

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
    chainID := "chainID_example" // string | ChainID (bech32-encoded)
    txHash := "txHash_example" // string | Transaction hash (hex-encoded)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDEvmReqidTxHashGet(context.Background(), chainID, txHash).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDEvmReqidTxHashGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDEvmReqidTxHashGet`: string
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDEvmReqidTxHashGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32-encoded) | 
**txHash** | **string** | Transaction hash (hex-encoded) | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDEvmReqidTxHashGetRequest struct via the builder pattern


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


## ChainChainIDRequestPost

> ChainChainIDRequestPost(ctx, chainID).Request(request).Execute()

Post an off-ledger request

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
    chainID := "chainID_example" // string | chainID
    request := *openapiclient.NewOffLedgerRequestBody() // OffLedgerRequestBody | Offledger Request encoded in base64. Optionally, the body can be the binary representation of the offledger request, but mime-type must be specified to \"application/octet-stream\" (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDRequestPost(context.Background(), chainID).Request(request).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDRequestPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | chainID | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDRequestPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**OffLedgerRequestBody**](OffLedgerRequestBody.md) | Offledger Request encoded in base64. Optionally, the body can be the binary representation of the offledger request, but mime-type must be specified to \&quot;application/octet-stream\&quot; | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ChainChainIDRequestReqIDReceiptGet

> RequestReceiptResponse ChainChainIDRequestReqIDReceiptGet(ctx, chainID, reqID).Execute()

Get the processing status of a given request in the node

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
    chainID := "chainID_example" // string | ChainID (bech32)
    reqID := "reqID_example" // string | Request ID

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDRequestReqIDReceiptGet(context.Background(), chainID, reqID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDRequestReqIDReceiptGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDRequestReqIDReceiptGet`: RequestReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDRequestReqIDReceiptGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32) | 
**reqID** | **string** | Request ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDRequestReqIDReceiptGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RequestReceiptResponse**](RequestReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ChainChainIDRequestReqIDWaitGet

> RequestReceiptResponse ChainChainIDRequestReqIDWaitGet(ctx, chainID, reqID).Params(params).Execute()

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
    chainID := "chainID_example" // string | ChainID (bech32)
    reqID := "reqID_example" // string | Request ID
    params := *openapiclient.NewWaitRequestProcessedParams() // WaitRequestProcessedParams | Optional parameters (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDRequestReqIDWaitGet(context.Background(), chainID, reqID).Params(params).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDRequestReqIDWaitGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDRequestReqIDWaitGet`: RequestReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDRequestReqIDWaitGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32) | 
**reqID** | **string** | Request ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDRequestReqIDWaitGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **params** | [**WaitRequestProcessedParams**](WaitRequestProcessedParams.md) | Optional parameters | 

### Return type

[**RequestReceiptResponse**](RequestReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ChainChainIDStateKeyGet

> []int32 ChainChainIDStateKeyGet(ctx, chainID, key).Execute()

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
    chainID := "chainID_example" // string | ChainID
    key := "key_example" // string | Key (hex-encoded)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDStateKeyGet(context.Background(), chainID, key).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDStateKeyGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChainChainIDStateKeyGet`: []int32
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ChainChainIDStateKeyGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID | 
**key** | **string** | Key (hex-encoded) | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDStateKeyGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

**[]int32**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ChainChainIDWsGet

> ChainChainIDWsGet(ctx, chainID).Execute()



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
    chainID := "chainID_example" // string | ChainID (bech32-encoded)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ChainChainIDWsGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ChainChainIDWsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32-encoded) | 

### Other Parameters

Other parameters are passed through a pointer to a apiChainChainIDWsGetRequest struct via the builder pattern


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


## InfoGet

> InfoResponse InfoGet(ctx).Execute()

Get information about the node

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
    resp, r, err := apiClient.PublicApi.InfoGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InfoGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InfoGet`: InfoResponse
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InfoGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInfoGetRequest struct via the builder pattern


### Return type

[**InfoResponse**](InfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

