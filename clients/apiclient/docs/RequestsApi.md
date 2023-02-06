# \RequestsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CallView**](RequestsApi.md#CallView) | **Post** /v2/requests/callview | Call a view function on a contract by Hname
[**GetReceipt**](RequestsApi.md#GetReceipt) | **Get** /v2/chains/{chainID}/receipts/{requestID} | Get a receipt from a request ID
[**OffLedger**](RequestsApi.md#OffLedger) | **Post** /v2/requests/offledger | Post an off-ledger request
[**WaitForRequest**](RequestsApi.md#WaitForRequest) | **Get** /v2/chains/{chainID}/requests/{requestID}/wait | Wait until the given request has been processed by the node



## CallView

> JSONDict CallView(ctx).ContractCallViewRequest(contractCallViewRequest).Execute()

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
    contractCallViewRequest := *openapiclient.NewContractCallViewRequest(*openapiclient.NewJSONDict(), "ChainId_example", "ContractHName_example", "ContractName_example", "FunctionHName_example", "FunctionName_example") // ContractCallViewRequest | Parameters

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.RequestsApi.CallView(context.Background()).ContractCallViewRequest(contractCallViewRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `RequestsApi.CallView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CallView`: JSONDict
    fmt.Fprintf(os.Stdout, "Response from `RequestsApi.CallView`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCallViewRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **contractCallViewRequest** | [**ContractCallViewRequest**](ContractCallViewRequest.md) | Parameters | 

### Return type

[**JSONDict**](JSONDict.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

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
    resp, r, err := apiClient.RequestsApi.GetReceipt(context.Background(), chainID, requestID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `RequestsApi.GetReceipt``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetReceipt`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `RequestsApi.GetReceipt`: %v\n", resp)
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

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OffLedger

> OffLedger(ctx).OffLedgerRequest(offLedgerRequest).Execute()

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
    offLedgerRequest := *openapiclient.NewOffLedgerRequest("ChainId_example", "Hex string") // OffLedgerRequest | Offledger request as JSON. Request encoded in Hex

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.RequestsApi.OffLedger(context.Background()).OffLedgerRequest(offLedgerRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `RequestsApi.OffLedger``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOffLedgerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offLedgerRequest** | [**OffLedgerRequest**](OffLedgerRequest.md) | Offledger request as JSON. Request encoded in Hex | 

### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WaitForRequest

> ReceiptResponse WaitForRequest(ctx, chainID, requestID).TimeoutSeconds(timeoutSeconds).Execute()

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
    timeoutSeconds := int32(56) // int32 | The timeout in seconds (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.RequestsApi.WaitForRequest(context.Background(), chainID, requestID).TimeoutSeconds(timeoutSeconds).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `RequestsApi.WaitForRequest``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `WaitForRequest`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `RequestsApi.WaitForRequest`: %v\n", resp)
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


 **timeoutSeconds** | **int32** | The timeout in seconds | 

### Return type

[**ReceiptResponse**](ReceiptResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

