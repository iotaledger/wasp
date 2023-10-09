# .RequestsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**offLedger**](RequestsApi.md#offLedger) | **POST** /v1/requests/offledger | Post an off-ledger request


# **offLedger**
> void offLedger(offLedgerRequest)


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .RequestsApi(configuration);

let body:.RequestsApiOffLedgerRequest = {
  // OffLedgerRequest | Offledger request as JSON. Request encoded in Hex
  offLedgerRequest: {
    chainId: "chainId_example",
    request: "Hex string",
  },
};

apiInstance.offLedger(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
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


