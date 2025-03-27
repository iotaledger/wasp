# .MetricsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getChainMessageMetrics**](MetricsApi.md#getChainMessageMetrics) | **GET** /v1/metrics/chain/messages | Get chain specific message metrics.
[**getChainPipeMetrics**](MetricsApi.md#getChainPipeMetrics) | **GET** /v1/metrics/chain/pipe | Get chain pipe event metrics.
[**getChainWorkflowMetrics**](MetricsApi.md#getChainWorkflowMetrics) | **GET** /v1/metrics/chain/workflow | Get chain workflow metrics.


# **getChainMessageMetrics**
> ChainMessageMetrics getChainMessageMetrics()


### Example


```typescript
import { createConfiguration, MetricsApi } from '';

const configuration = createConfiguration();
const apiInstance = new MetricsApi(configuration);

const request = {};

const data = await apiInstance.getChainMessageMetrics(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**ChainMessageMetrics**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all available metrics. |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | Chain not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getChainPipeMetrics**
> ConsensusPipeMetrics getChainPipeMetrics()


### Example


```typescript
import { createConfiguration, MetricsApi } from '';

const configuration = createConfiguration();
const apiInstance = new MetricsApi(configuration);

const request = {};

const data = await apiInstance.getChainPipeMetrics(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**ConsensusPipeMetrics**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all available metrics. |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | Chain not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getChainWorkflowMetrics**
> ConsensusWorkflowMetrics getChainWorkflowMetrics()


### Example


```typescript
import { createConfiguration, MetricsApi } from '';

const configuration = createConfiguration();
const apiInstance = new MetricsApi(configuration);

const request = {};

const data = await apiInstance.getChainWorkflowMetrics(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**ConsensusWorkflowMetrics**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all available metrics. |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | Chain not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


