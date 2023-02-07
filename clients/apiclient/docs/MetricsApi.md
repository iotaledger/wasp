# \MetricsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetChainMetrics**](MetricsApi.md#GetChainMetrics) | **Get** /metrics/chain/{chainID} | Get chain specific metrics.
[**GetChainPipeMetrics**](MetricsApi.md#GetChainPipeMetrics) | **Get** /metrics/chain/{chainID}/pipe | Get chain pipe event metrics.
[**GetChainWorkflowMetrics**](MetricsApi.md#GetChainWorkflowMetrics) | **Get** /metrics/chain/{chainID}/workflow | Get chain workflow metrics.
[**GetL1Metrics**](MetricsApi.md#GetL1Metrics) | **Get** /metrics/l1 | Get accumulated metrics.



## GetChainMetrics

> ChainMetrics GetChainMetrics(ctx, chainID).Execute()

Get chain specific metrics.

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
    resp, r, err := apiClient.MetricsApi.GetChainMetrics(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.GetChainMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetChainMetrics`: ChainMetrics
    fmt.Fprintf(os.Stdout, "Response from `MetricsApi.GetChainMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ChainMetrics**](ChainMetrics.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetChainPipeMetrics

> ConsensusPipeMetrics GetChainPipeMetrics(ctx, chainID).Execute()

Get chain pipe event metrics.

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
    resp, r, err := apiClient.MetricsApi.GetChainPipeMetrics(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.GetChainPipeMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetChainPipeMetrics`: ConsensusPipeMetrics
    fmt.Fprintf(os.Stdout, "Response from `MetricsApi.GetChainPipeMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainPipeMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ConsensusPipeMetrics**](ConsensusPipeMetrics.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetChainWorkflowMetrics

> ConsensusWorkflowMetrics GetChainWorkflowMetrics(ctx, chainID).Execute()

Get chain workflow metrics.

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
    resp, r, err := apiClient.MetricsApi.GetChainWorkflowMetrics(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.GetChainWorkflowMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetChainWorkflowMetrics`: ConsensusWorkflowMetrics
    fmt.Fprintf(os.Stdout, "Response from `MetricsApi.GetChainWorkflowMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainWorkflowMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ConsensusWorkflowMetrics**](ConsensusWorkflowMetrics.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetL1Metrics

> ChainMetrics GetL1Metrics(ctx).Execute()

Get accumulated metrics.

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
    resp, r, err := apiClient.MetricsApi.GetL1Metrics(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.GetL1Metrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetL1Metrics`: ChainMetrics
    fmt.Fprintf(os.Stdout, "Response from `MetricsApi.GetL1Metrics`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetL1MetricsRequest struct via the builder pattern


### Return type

[**ChainMetrics**](ChainMetrics.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

