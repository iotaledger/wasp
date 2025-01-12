# .MetricsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getChainMessageMetrics**](MetricsApi.md#getChainMessageMetrics) | **GET** /v1/metrics/chain/{chainID}/messages | Get chain specific message metrics.
[**getChainPipeMetrics**](MetricsApi.md#getChainPipeMetrics) | **GET** /v1/metrics/chain/{chainID}/pipe | Get chain pipe event metrics.
[**getChainWorkflowMetrics**](MetricsApi.md#getChainWorkflowMetrics) | **GET** /v1/metrics/chain/{chainID}/workflow | Get chain workflow metrics.
[**getNodeMessageMetrics**](MetricsApi.md#getNodeMessageMetrics) | **GET** /v1/metrics/node/messages | Get accumulated message metrics.


# **getChainMessageMetrics**
> ChainMessageMetrics getChainMessageMetrics()


### Example


```typescript
import { createConfiguration, MetricsApi } from '';
import type { MetricsApiGetChainMessageMetricsRequest } from '';

const configuration = createConfiguration();
const apiInstance = new MetricsApi(configuration);

const request: MetricsApiGetChainMessageMetricsRequest = {
    // ChainID (Hex Address)
  chainID: "chainID_example",
};

const data = await apiInstance.getChainMessageMetrics(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Hex Address) | defaults to undefined


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
import type { MetricsApiGetChainPipeMetricsRequest } from '';

const configuration = createConfiguration();
const apiInstance = new MetricsApi(configuration);

const request: MetricsApiGetChainPipeMetricsRequest = {
    // ChainID (Hex Address)
  chainID: "chainID_example",
};

const data = await apiInstance.getChainPipeMetrics(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Hex Address) | defaults to undefined


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
import type { MetricsApiGetChainWorkflowMetricsRequest } from '';

const configuration = createConfiguration();
const apiInstance = new MetricsApi(configuration);

const request: MetricsApiGetChainWorkflowMetricsRequest = {
    // ChainID (Hex Address)
  chainID: "chainID_example",
};

const data = await apiInstance.getChainWorkflowMetrics(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Hex Address) | defaults to undefined


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

# **getNodeMessageMetrics**
> NodeMessageMetrics getNodeMessageMetrics()


### Example


```typescript
import { createConfiguration, MetricsApi } from '';

const configuration = createConfiguration();
const apiInstance = new MetricsApi(configuration);

const request = {};

const data = await apiInstance.getNodeMessageMetrics(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**NodeMessageMetrics**

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

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


