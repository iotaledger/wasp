# \MetricsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetChainMessageMetrics**](MetricsAPI.md#GetChainMessageMetrics) | **Get** /v1/metrics/chain/messages | Get chain specific message metrics.
[**GetChainPipeMetrics**](MetricsAPI.md#GetChainPipeMetrics) | **Get** /v1/metrics/chain/pipe | Get chain pipe event metrics.
[**GetChainWorkflowMetrics**](MetricsAPI.md#GetChainWorkflowMetrics) | **Get** /v1/metrics/chain/workflow | Get chain workflow metrics.



## GetChainMessageMetrics

> ChainMessageMetrics GetChainMessageMetrics(ctx).Execute()

Get chain specific message metrics.

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MetricsAPI.GetChainMessageMetrics(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricsAPI.GetChainMessageMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetChainMessageMetrics`: ChainMessageMetrics
	fmt.Fprintf(os.Stdout, "Response from `MetricsAPI.GetChainMessageMetrics`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainMessageMetricsRequest struct via the builder pattern


### Return type

[**ChainMessageMetrics**](ChainMessageMetrics.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetChainPipeMetrics

> ConsensusPipeMetrics GetChainPipeMetrics(ctx).Execute()

Get chain pipe event metrics.

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MetricsAPI.GetChainPipeMetrics(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricsAPI.GetChainPipeMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetChainPipeMetrics`: ConsensusPipeMetrics
	fmt.Fprintf(os.Stdout, "Response from `MetricsAPI.GetChainPipeMetrics`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainPipeMetricsRequest struct via the builder pattern


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

> ConsensusWorkflowMetrics GetChainWorkflowMetrics(ctx).Execute()

Get chain workflow metrics.

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MetricsAPI.GetChainWorkflowMetrics(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricsAPI.GetChainWorkflowMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetChainWorkflowMetrics`: ConsensusWorkflowMetrics
	fmt.Fprintf(os.Stdout, "Response from `MetricsAPI.GetChainWorkflowMetrics`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetChainWorkflowMetricsRequest struct via the builder pattern


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

