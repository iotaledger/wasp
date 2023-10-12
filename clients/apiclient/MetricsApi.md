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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .MetricsApi(configuration);

let body:.MetricsApiGetChainMessageMetricsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.getChainMessageMetrics(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .MetricsApi(configuration);

let body:.MetricsApiGetChainPipeMetricsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.getChainPipeMetrics(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .MetricsApi(configuration);

let body:.MetricsApiGetChainWorkflowMetricsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.getChainWorkflowMetrics(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .MetricsApi(configuration);

let body:any = {};

apiInstance.getNodeMessageMetrics(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
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


