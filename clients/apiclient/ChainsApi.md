# .ChainsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**activateChain**](ChainsApi.md#activateChain) | **POST** /v1/chains/{chainID}/activate | Activate a chain
[**addAccessNode**](ChainsApi.md#addAccessNode) | **PUT** /v1/chains/{chainID}/access-node/{peer} | Configure a trusted node to be an access node.
[**callView**](ChainsApi.md#callView) | **POST** /v1/chains/{chainID}/callview | Call a view function on a contract by Hname
[**deactivateChain**](ChainsApi.md#deactivateChain) | **POST** /v1/chains/{chainID}/deactivate | Deactivate a chain
[**dumpAccounts**](ChainsApi.md#dumpAccounts) | **POST** /v1/chains/{chainID}/dump-accounts | dump accounts information into a humanly-readable format
[**estimateGasOffledger**](ChainsApi.md#estimateGasOffledger) | **POST** /v1/chains/{chainID}/estimategas-offledger | Estimates gas for a given off-ledger ISC request
[**estimateGasOnledger**](ChainsApi.md#estimateGasOnledger) | **POST** /v1/chains/{chainID}/estimategas-onledger | Estimates gas for a given on-ledger ISC request
[**getChainInfo**](ChainsApi.md#getChainInfo) | **GET** /v1/chains/{chainID} | Get information about a specific chain
[**getChains**](ChainsApi.md#getChains) | **GET** /v1/chains | Get a list of all chains
[**getCommitteeInfo**](ChainsApi.md#getCommitteeInfo) | **GET** /v1/chains/{chainID}/committee | Get information about the deployed committee
[**getContracts**](ChainsApi.md#getContracts) | **GET** /v1/chains/{chainID}/contracts | Get all available chain contracts
[**getMempoolContents**](ChainsApi.md#getMempoolContents) | **GET** /v1/chains/{chainID}/mempool | Get the contents of the mempool.
[**getReceipt**](ChainsApi.md#getReceipt) | **GET** /v1/chains/{chainID}/receipts/{requestID} | Get a receipt from a request ID
[**getStateValue**](ChainsApi.md#getStateValue) | **GET** /v1/chains/{chainID}/state/{stateKey} | Fetch the raw value associated with the given key in the chain state
[**removeAccessNode**](ChainsApi.md#removeAccessNode) | **DELETE** /v1/chains/{chainID}/access-node/{peer} | Remove an access node.
[**setChainRecord**](ChainsApi.md#setChainRecord) | **POST** /v1/chains/{chainID}/chainrecord | Sets the chain record.
[**v1ChainsChainIDEvmPost**](ChainsApi.md#v1ChainsChainIDEvmPost) | **POST** /v1/chains/{chainID}/evm | Ethereum JSON-RPC
[**v1ChainsChainIDEvmWsGet**](ChainsApi.md#v1ChainsChainIDEvmWsGet) | **GET** /v1/chains/{chainID}/evm/ws | Ethereum JSON-RPC (Websocket transport)
[**waitForRequest**](ChainsApi.md#waitForRequest) | **GET** /v1/chains/{chainID}/requests/{requestID}/wait | Wait until the given request has been processed by the node


# **activateChain**
> void activateChain()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiActivateChainRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.activateChain(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiAddAccessNodeRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Name or PubKey (hex) of the trusted peer
  peer: "peer_example",
};

apiInstance.addAccessNode(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
> JSONDict callView(contractCallViewRequest)

Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.

### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiCallViewRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // ContractCallViewRequest | Parameters
  contractCallViewRequest: {
    arguments: {
      items: [
        {
          key: "key_example",
          value: "value_example",
        },
      ],
    },
    block: "block_example",
    contractHName: "contractHName_example",
    contractName: "contractName_example",
    functionHName: "functionHName_example",
    functionName: "functionName_example",
  },
};

apiInstance.callView(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **contractCallViewRequest** | **ContractCallViewRequest**| Parameters |
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


### Return type

**JSONDict**

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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiDeactivateChainRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.deactivateChain(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiDumpAccountsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.dumpAccounts(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiEstimateGasOffledgerRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // EstimateGasRequestOffledger | Request
  request: {
    fromAddress: "fromAddress_example",
    requestBytes: "requestBytes_example",
  },
};

apiInstance.estimateGasOffledger(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | **EstimateGasRequestOffledger**| Request |
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
> ReceiptResponse estimateGasOnledger(request)


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiEstimateGasOnledgerRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // EstimateGasRequestOnledger | Request
  request: {
    outputBytes: "outputBytes_example",
  },
};

apiInstance.estimateGasOnledger(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | **EstimateGasRequestOnledger**| Request |
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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

# **getChainInfo**
> ChainInfoResponse getChainInfo()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiGetChainInfoRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.getChainInfo(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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

# **getChains**
> Array<ChainInfoResponse> getChains()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:any = {};

apiInstance.getChains(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters
This endpoint does not need any parameter.


### Return type

**Array<ChainInfoResponse>**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all available chains |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getCommitteeInfo**
> CommitteeInfoResponse getCommitteeInfo()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiGetCommitteeInfoRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.getCommitteeInfo(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiGetContractsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.getContracts(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiGetMempoolContentsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.getMempoolContents(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiGetReceiptRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | RequestID (Hex)
  requestID: "requestID_example",
};

apiInstance.getReceipt(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiGetStateValueRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | State Key (Hex)
  stateKey: "stateKey_example",
};

apiInstance.getStateValue(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiRemoveAccessNodeRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Name or PubKey (hex) of the trusted peer
  peer: "peer_example",
};

apiInstance.removeAccessNode(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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

# **setChainRecord**
> void setChainRecord(chainRecord)


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiSetChainRecordRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // ChainRecord | Chain Record
  chainRecord: {
    accessNodes: [
      "accessNodes_example",
    ],
    isActive: true,
  },
};

apiInstance.setChainRecord(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainRecord** | **ChainRecord**| Chain Record |
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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

# **v1ChainsChainIDEvmPost**
> v1ChainsChainIDEvmPost()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiV1ChainsChainIDEvmPostRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.v1ChainsChainIDEvmPost(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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

# **v1ChainsChainIDEvmWsGet**
> v1ChainsChainIDEvmWsGet()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiV1ChainsChainIDEvmWsGetRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
};

apiInstance.v1ChainsChainIDEvmWsGet(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined


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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .ChainsApi(configuration);

let body:.ChainsApiWaitForRequestRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | RequestID (Hex)
  requestID: "requestID_example",
  // number | The timeout in seconds, maximum 60s (optional)
  timeoutSeconds: 1,
  // boolean | Wait for the block to be confirmed on L1 (optional)
  waitForL1Confirmation: true,
};

apiInstance.waitForRequest(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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


