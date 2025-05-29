# .CorecontractsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**accountsGetAccountBalance**](CorecontractsApi.md#accountsGetAccountBalance) | **GET** /v1/chain/core/accounts/account/{agentID}/balance | Get all assets belonging to an account
[**accountsGetAccountNonce**](CorecontractsApi.md#accountsGetAccountNonce) | **GET** /v1/chain/core/accounts/account/{agentID}/nonce | Get the current nonce of an account
[**accountsGetTotalAssets**](CorecontractsApi.md#accountsGetTotalAssets) | **GET** /v1/chain/core/accounts/total_assets | Get all stored assets
[**blocklogGetBlockInfo**](CorecontractsApi.md#blocklogGetBlockInfo) | **GET** /v1/chain/core/blocklog/blocks/{blockIndex} | Get the block info of a certain block index
[**blocklogGetControlAddresses**](CorecontractsApi.md#blocklogGetControlAddresses) | **GET** /v1/chain/core/blocklog/controladdresses | Get the control addresses
[**blocklogGetEventsOfBlock**](CorecontractsApi.md#blocklogGetEventsOfBlock) | **GET** /v1/chain/core/blocklog/events/block/{blockIndex} | Get events of a block
[**blocklogGetEventsOfLatestBlock**](CorecontractsApi.md#blocklogGetEventsOfLatestBlock) | **GET** /v1/chain/core/blocklog/events/block/latest | Get events of the latest block
[**blocklogGetEventsOfRequest**](CorecontractsApi.md#blocklogGetEventsOfRequest) | **GET** /v1/chain/core/blocklog/events/request/{requestID} | Get events of a request
[**blocklogGetLatestBlockInfo**](CorecontractsApi.md#blocklogGetLatestBlockInfo) | **GET** /v1/chain/core/blocklog/blocks/latest | Get the block info of the latest block
[**blocklogGetRequestIDsForBlock**](CorecontractsApi.md#blocklogGetRequestIDsForBlock) | **GET** /v1/chain/core/blocklog/blocks/{blockIndex}/requestids | Get the request ids for a certain block index
[**blocklogGetRequestIDsForLatestBlock**](CorecontractsApi.md#blocklogGetRequestIDsForLatestBlock) | **GET** /v1/chain/core/blocklog/blocks/latest/requestids | Get the request ids for the latest block
[**blocklogGetRequestIsProcessed**](CorecontractsApi.md#blocklogGetRequestIsProcessed) | **GET** /v1/chain/core/blocklog/requests/{requestID}/is_processed | Get the request processing status
[**blocklogGetRequestReceipt**](CorecontractsApi.md#blocklogGetRequestReceipt) | **GET** /v1/chain/core/blocklog/requests/{requestID} | Get the receipt of a certain request id
[**blocklogGetRequestReceiptsOfBlock**](CorecontractsApi.md#blocklogGetRequestReceiptsOfBlock) | **GET** /v1/chain/core/blocklog/blocks/{blockIndex}/receipts | Get all receipts of a certain block
[**blocklogGetRequestReceiptsOfLatestBlock**](CorecontractsApi.md#blocklogGetRequestReceiptsOfLatestBlock) | **GET** /v1/chain/core/blocklog/blocks/latest/receipts | Get all receipts of the latest block
[**errorsGetErrorMessageFormat**](CorecontractsApi.md#errorsGetErrorMessageFormat) | **GET** /v1/chain/core/errors/{contractHname}/message/{errorID} | Get the error message format of a specific error id
[**governanceGetChainAdmin**](CorecontractsApi.md#governanceGetChainAdmin) | **GET** /v1/chain/core/governance/chainadmin | Get the chain admin
[**governanceGetChainInfo**](CorecontractsApi.md#governanceGetChainInfo) | **GET** /v1/chain/core/governance/chaininfo | Get the chain info


# **accountsGetAccountBalance**
> AssetsResponse accountsGetAccountBalance()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiAccountsGetAccountBalanceRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiAccountsGetAccountBalanceRequest = {
    // AgentID (Hex Address for L1 accounts | Hex for EVM)
  agentID: "agentID_example",
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.accountsGetAccountBalance(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **agentID** | [**string**] | AgentID (Hex Address for L1 accounts | Hex for EVM) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**AssetsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | All assets belonging to an account |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **accountsGetAccountNonce**
> AccountNonceResponse accountsGetAccountNonce()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiAccountsGetAccountNonceRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiAccountsGetAccountNonceRequest = {
    // AgentID (Hex Address for L1 accounts | Hex for EVM)
  agentID: "agentID_example",
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.accountsGetAccountNonce(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **agentID** | [**string**] | AgentID (Hex Address for L1 accounts | Hex for EVM) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**AccountNonceResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The current nonce of an account |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **accountsGetTotalAssets**
> AssetsResponse accountsGetTotalAssets()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiAccountsGetTotalAssetsRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiAccountsGetTotalAssetsRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.accountsGetTotalAssets(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**AssetsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | All stored assets |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetBlockInfo**
> BlockInfoResponse blocklogGetBlockInfo()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetBlockInfoRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetBlockInfoRequest = {
    // BlockIndex (uint32)
  blockIndex: 1,
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetBlockInfo(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **blockIndex** | [**number**] | BlockIndex (uint32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**BlockInfoResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The block info |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetControlAddresses**
> ControlAddressesResponse blocklogGetControlAddresses()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetControlAddressesRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetControlAddressesRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetControlAddresses(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**ControlAddressesResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The chain info |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetEventsOfBlock**
> EventsResponse blocklogGetEventsOfBlock()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetEventsOfBlockRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetEventsOfBlockRequest = {
    // BlockIndex (uint32)
  blockIndex: 1,
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetEventsOfBlock(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **blockIndex** | [**number**] | BlockIndex (uint32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**EventsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The events |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetEventsOfLatestBlock**
> EventsResponse blocklogGetEventsOfLatestBlock()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetEventsOfLatestBlockRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetEventsOfLatestBlockRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetEventsOfLatestBlock(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**EventsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The receipts |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetEventsOfRequest**
> EventsResponse blocklogGetEventsOfRequest()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetEventsOfRequestRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetEventsOfRequestRequest = {
    // RequestID (Hex)
  requestID: "requestID_example",
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetEventsOfRequest(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestID** | [**string**] | RequestID (Hex) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**EventsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The events |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetLatestBlockInfo**
> BlockInfoResponse blocklogGetLatestBlockInfo()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetLatestBlockInfoRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetLatestBlockInfoRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetLatestBlockInfo(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**BlockInfoResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The block info |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetRequestIDsForBlock**
> RequestIDsResponse blocklogGetRequestIDsForBlock()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetRequestIDsForBlockRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetRequestIDsForBlockRequest = {
    // BlockIndex (uint32)
  blockIndex: 1,
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetRequestIDsForBlock(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **blockIndex** | [**number**] | BlockIndex (uint32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**RequestIDsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of request ids (ISCRequestID[]) |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetRequestIDsForLatestBlock**
> RequestIDsResponse blocklogGetRequestIDsForLatestBlock()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetRequestIDsForLatestBlockRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetRequestIDsForLatestBlockRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetRequestIDsForLatestBlock(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**RequestIDsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of request ids (ISCRequestID[]) |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetRequestIsProcessed**
> RequestProcessedResponse blocklogGetRequestIsProcessed()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetRequestIsProcessedRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetRequestIsProcessedRequest = {
    // RequestID (Hex)
  requestID: "requestID_example",
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetRequestIsProcessed(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestID** | [**string**] | RequestID (Hex) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**RequestProcessedResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The processing result |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetRequestReceipt**
> ReceiptResponse blocklogGetRequestReceipt()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetRequestReceiptRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetRequestReceiptRequest = {
    // RequestID (Hex)
  requestID: "requestID_example",
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetRequestReceipt(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestID** | [**string**] | RequestID (Hex) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**ReceiptResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The receipt |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetRequestReceiptsOfBlock**
> Array<ReceiptResponse> blocklogGetRequestReceiptsOfBlock()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetRequestReceiptsOfBlockRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetRequestReceiptsOfBlockRequest = {
    // BlockIndex (uint32)
  blockIndex: 1,
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetRequestReceiptsOfBlock(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **blockIndex** | [**number**] | BlockIndex (uint32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**Array<ReceiptResponse>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The receipts |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetRequestReceiptsOfLatestBlock**
> Array<ReceiptResponse> blocklogGetRequestReceiptsOfLatestBlock()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiBlocklogGetRequestReceiptsOfLatestBlockRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiBlocklogGetRequestReceiptsOfLatestBlockRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.blocklogGetRequestReceiptsOfLatestBlock(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**Array<ReceiptResponse>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The receipts |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **errorsGetErrorMessageFormat**
> ErrorMessageFormatResponse errorsGetErrorMessageFormat()


### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiErrorsGetErrorMessageFormatRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiErrorsGetErrorMessageFormatRequest = {
    // ChainID (Hex Address)
  chainID: "chainID_example",
    // Contract (Hname as Hex)
  contractHname: "contractHname_example",
    // Error Id (uint16)
  errorID: 1,
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.errorsGetErrorMessageFormat(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Hex Address) | defaults to undefined
 **contractHname** | [**string**] | Contract (Hname as Hex) | defaults to undefined
 **errorID** | [**number**] | Error Id (uint16) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**ErrorMessageFormatResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The error message format |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **governanceGetChainAdmin**
> GovChainAdminResponse governanceGetChainAdmin()

Returns the chain admin

### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiGovernanceGetChainAdminRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiGovernanceGetChainAdminRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.governanceGetChainAdmin(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**GovChainAdminResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The chain admin |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **governanceGetChainInfo**
> GovChainInfoResponse governanceGetChainInfo()

If you are using the common API functions, you most likely rather want to use \'/v1/chains/:chainID\' to get information about a chain.

### Example


```typescript
import { createConfiguration, CorecontractsApi } from '';
import type { CorecontractsApiGovernanceGetChainInfoRequest } from '';

const configuration = createConfiguration();
const apiInstance = new CorecontractsApi(configuration);

const request: CorecontractsApiGovernanceGetChainInfoRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.governanceGetChainInfo(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**GovChainInfoResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The chain info |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


