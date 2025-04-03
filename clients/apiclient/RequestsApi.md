# .RequestsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**offLedger**](RequestsApi.md#offLedger) | **POST** /v1/requests/offledger | Post an off-ledger request


# **offLedger**
> void offLedger(offLedgerRequest)


### Example


```typescript
import { createConfiguration, RequestsApi } from '';
import type { RequestsApiOffLedgerRequest } from '';

const configuration = createConfiguration();
const apiInstance = new RequestsApi(configuration);

const request: RequestsApiOffLedgerRequest = {
    // Offledger request as JSON. Request encoded in Hex
  offLedgerRequest: {
    request: "Hex string",
  },
};

const data = await apiInstance.offLedger(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offLedgerRequest** | **OffLedgerRequest**| Offledger request as JSON. Request encoded in Hex |


### Return type

**void**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**202** | Request submitted |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


