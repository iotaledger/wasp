# .ChainsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**activateChain**](ChainsApi.md#activateChain) | **POST** /v1/chain/activate/{chainID} | Activate a chain
[**addAccessNode**](ChainsApi.md#addAccessNode) | **PUT** /v1/chain/access-node/{peer} | Configure a trusted node to be an access node.
[**callView**](ChainsApi.md#callView) | **POST** /v1/chain/callview | Call a view function on a contract by Hname
[**deactivateChain**](ChainsApi.md#deactivateChain) | **POST** /v1/chain/deactivate | Deactivate a chain
[**dumpAccounts**](ChainsApi.md#dumpAccounts) | **POST** /v1/chain/dump-accounts | dump accounts information into a humanly-readable format
[**estimateGasOffledger**](ChainsApi.md#estimateGasOffledger) | **POST** /v1/chain/estimategas-offledger | Estimates gas for a given off-ledger ISC request
[**estimateGasOnledger**](ChainsApi.md#estimateGasOnledger) | **POST** /v1/chain/estimategas-onledger | Estimates gas for a given on-ledger ISC request
[**getChainInfo**](ChainsApi.md#getChainInfo) | **GET** /v1/chain | Get information about a specific chain
[**getCommitteeInfo**](ChainsApi.md#getCommitteeInfo) | **GET** /v1/chain/committee | Get information about the deployed committee
[**getContracts**](ChainsApi.md#getContracts) | **GET** /v1/chain/contracts | Get all available chain contracts
[**getMempoolContents**](ChainsApi.md#getMempoolContents) | **GET** /v1/chain/mempool | Get the contents of the mempool.
[**getReceipt**](ChainsApi.md#getReceipt) | **GET** /v1/chain/receipts/{requestID} | Get a receipt from a request ID
[**getStateValue**](ChainsApi.md#getStateValue) | **GET** /v1/chain/state/{stateKey} | Fetch the raw value associated with the given key in the chain state
[**removeAccessNode**](ChainsApi.md#removeAccessNode) | **DELETE** /v1/chain/access-node/{peer} | Remove an access node.
[**rotateChain**](ChainsApi.md#rotateChain) | **POST** /v1/chain/rotate | Rotate a chain
[**setChainRecord**](ChainsApi.md#setChainRecord) | **POST** /v1/chain/chainrecord/{chainID} | Sets the chain record.
[**v1ChainEvmPost**](ChainsApi.md#v1ChainEvmPost) | **POST** /v1/chain/evm | Ethereum JSON-RPC
[**v1ChainEvmWsGet**](ChainsApi.md#v1ChainEvmWsGet) | **GET** /v1/chain/evm/ws | Ethereum JSON-RPC (Websocket transport)
[**waitForRequest**](ChainsApi.md#waitForRequest) | **GET** /v1/chain/requests/{requestID}/wait | Wait until the given request has been processed by the node


# **activateChain**
> void activateChain()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiActivateChainRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiActivateChainRequest = {
    // ChainID (Hex Address)
  chainID: "chainID_example",
};

const data = await apiInstance.activateChain(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Hex Address) | defaults to undefined


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Chain was successfully activated |  -  |
**304** | Chain was not activated |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **addAccessNode**
> void addAccessNode()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiAddAccessNodeRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiAddAccessNodeRequest = {
    // Name or PubKey (hex) of the trusted peer
  peer: "peer_example",
};

const data = await apiInstance.addAccessNode(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **peer** | [**string**] | Name or PubKey (hex) of the trusted peer | defaults to undefined


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Access node was successfully added |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **callView**
> Array<string> callView(contractCallViewRequest)

Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.

### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiCallViewRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiCallViewRequest = {
    // Parameters
  contractCallViewRequest: {
    arguments: [
      "arguments_example",
    ],
    block: "block_example",
    contractHName: "contractHName_example",
    contractName: "contractName_example",
    functionHName: "functionHName_example",
    functionName: "functionName_example",
  },
};

const data = await apiInstance.callView(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **contractCallViewRequest** | **ContractCallViewRequest**| Parameters |


### Return type

**Array<string>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Result |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **deactivateChain**
> void deactivateChain()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request = {};

const data = await apiInstance.deactivateChain(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Chain was successfully deactivated |  -  |
**304** | Chain was not deactivated |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **dumpAccounts**
> void dumpAccounts()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request = {};

const data = await apiInstance.dumpAccounts(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Accounts dump will be produced |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **estimateGasOffledger**
> ReceiptResponse estimateGasOffledger(request)


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiEstimateGasOffledgerRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiEstimateGasOffledgerRequest = {
    // Request
  request: {
    fromAddress: "fromAddress_example",
    requestBytes: "requestBytes_example",
  },
};

const data = await apiInstance.estimateGasOffledger(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | **EstimateGasRequestOffledger**| Request |


### Return type

**ReceiptResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ReceiptResponse |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **estimateGasOnledger**
> OnLedgerEstimationResponse estimateGasOnledger(request)


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiEstimateGasOnledgerRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiEstimateGasOnledgerRequest = {
    // Request
  request: {
    transactionBytes: "transactionBytes_example",
  },
};

const data = await apiInstance.estimateGasOnledger(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | **EstimateGasRequestOnledger**| Request |


### Return type

**OnLedgerEstimationResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ReceiptResponse |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getChainInfo**
> ChainInfoResponse getChainInfo()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiGetChainInfoRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiGetChainInfoRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.getChainInfo(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**ChainInfoResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Information about a specific chain |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getCommitteeInfo**
> CommitteeInfoResponse getCommitteeInfo()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiGetCommitteeInfoRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiGetCommitteeInfoRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.getCommitteeInfo(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**CommitteeInfoResponse**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all nodes tied to the chain |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getContracts**
> Array<ContractInfoResponse> getContracts()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiGetContractsRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiGetContractsRequest = {
    // Block index or trie root (optional)
  block: "block_example",
};

const data = await apiInstance.getContracts(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**Array<ContractInfoResponse>**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all available contracts |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getMempoolContents**
> Array<number> getMempoolContents()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request = {};

const data = await apiInstance.getMempoolContents(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**Array<number>**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | stream of JSON representation of the requests in the mempool |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getReceipt**
> ReceiptResponse getReceipt()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiGetReceiptRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiGetReceiptRequest = {
    // RequestID (Hex)
  requestID: "requestID_example",
};

const data = await apiInstance.getReceipt(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestID** | [**string**] | RequestID (Hex) | defaults to undefined


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
**200** | ReceiptResponse |  -  |
**404** | Chain or request id not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getStateValue**
> StateResponse getStateValue()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiGetStateValueRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiGetStateValueRequest = {
    // State Key (Hex)
  stateKey: "stateKey_example",
};

const data = await apiInstance.getStateValue(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **stateKey** | [**string**] | State Key (Hex) | defaults to undefined


### Return type

**StateResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Result |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **removeAccessNode**
> void removeAccessNode()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiRemoveAccessNodeRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiRemoveAccessNodeRequest = {
    // Name or PubKey (hex) of the trusted peer
  peer: "peer_example",
};

const data = await apiInstance.removeAccessNode(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **peer** | [**string**] | Name or PubKey (hex) of the trusted peer | defaults to undefined


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Access node was successfully removed |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **rotateChain**
> void rotateChain()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiRotateChainRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiRotateChainRequest = {
    // RotateRequest (optional)
  rotateRequest: {
    rotateToAddress: "rotateToAddress_example",
  },
};

const data = await apiInstance.rotateChain(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **rotateRequest** | **RotateChainRequest**| RotateRequest |


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Chain rotation was requested |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **setChainRecord**
> void setChainRecord(chainRecord)


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiSetChainRecordRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiSetChainRecordRequest = {
    // ChainID (Hex Address)
  chainID: "chainID_example",
    // Chain Record
  chainRecord: {
    accessNodes: [
      "accessNodes_example",
    ],
    isActive: true,
  },
};

const data = await apiInstance.setChainRecord(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainRecord** | **ChainRecord**| Chain Record |
 **chainID** | [**string**] | ChainID (Hex Address) | defaults to undefined


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | Chain record was saved |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **v1ChainEvmPost**
> v1ChainEvmPost()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request = {};

const data = await apiInstance.v1ChainEvmPost(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**0** | successful operation |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **v1ChainEvmWsGet**
> v1ChainEvmWsGet()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request = {};

const data = await apiInstance.v1ChainEvmWsGet(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**0** | successful operation |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **waitForRequest**
> ReceiptResponse waitForRequest()


### Example


```typescript
import { createConfiguration, ChainsApi } from '';
import type { ChainsApiWaitForRequestRequest } from '';

const configuration = createConfiguration();
const apiInstance = new ChainsApi(configuration);

const request: ChainsApiWaitForRequestRequest = {
    // RequestID (Hex)
  requestID: "requestID_example",
    // The timeout in seconds, maximum 60s (optional)
  timeoutSeconds: 1,
    // Wait for the block to be confirmed on L1 (optional)
  waitForL1Confirmation: true,
};

const data = await apiInstance.waitForRequest(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestID** | [**string**] | RequestID (Hex) | defaults to undefined
 **timeoutSeconds** | [**number**] | The timeout in seconds, maximum 60s | (optional) defaults to undefined
 **waitForL1Confirmation** | [**boolean**] | Wait for the block to be confirmed on L1 | (optional) defaults to undefined


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
**200** | The request receipt |  -  |
**404** | The chain or request id not found |  -  |
**408** | The waiting time has reached the defined limit |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


