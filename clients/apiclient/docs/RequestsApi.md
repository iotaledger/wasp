# \RequestsApi

All URIs are relative to *http://localhost:9090*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CallView**](RequestsApi.md#CallView) | **Post** /requests/callview | Call a view function on a contract by Hname
[**GetReceipt**](RequestsApi.md#GetReceipt) | **Get** /chains/{chainID}/receipts/{requestID} | Get a receipt from a request ID
[**OffLedger**](RequestsApi.md#OffLedger) | **Post** /requests/offledger | Post an off-ledger request
[**WaitForTransaction**](RequestsApi.md#WaitForTransaction) | **Get** /chains/{chainID}/requests/{requestID}/wait | Wait until the given request has been processed by the node



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

No authorization required

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

No authorization required

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

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WaitForTransaction

> ReceiptResponse WaitForTransaction(ctx, chainID, requestID).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.RequestsApi.WaitForTransaction(context.Background(), chainID, requestID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `RequestsApi.WaitForTransaction``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `WaitForTransaction`: ReceiptResponse
    fmt.Fprintf(os.Stdout, "Response from `RequestsApi.WaitForTransaction`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiWaitForTransactionRequest struct via the builder pattern


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

