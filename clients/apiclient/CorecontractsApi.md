# .CorecontractsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**accountsGetAccountBalance**](CorecontractsApi.md#accountsGetAccountBalance) | **GET** /v1/chains/{chainID}/core/accounts/account/{agentID}/balance | Get all assets belonging to an account
[**accountsGetAccountFoundries**](CorecontractsApi.md#accountsGetAccountFoundries) | **GET** /v1/chains/{chainID}/core/accounts/account/{agentID}/foundries | Get all foundries owned by an account
[**accountsGetAccountNFTIDs**](CorecontractsApi.md#accountsGetAccountNFTIDs) | **GET** /v1/chains/{chainID}/core/accounts/account/{agentID}/nfts | Get all NFT ids belonging to an account
[**accountsGetAccountNonce**](CorecontractsApi.md#accountsGetAccountNonce) | **GET** /v1/chains/{chainID}/core/accounts/account/{agentID}/nonce | Get the current nonce of an account
[**accountsGetFoundryOutput**](CorecontractsApi.md#accountsGetFoundryOutput) | **GET** /v1/chains/{chainID}/core/accounts/foundry_output/{serialNumber} | Get the foundry output
[**accountsGetNFTData**](CorecontractsApi.md#accountsGetNFTData) | **GET** /v1/chains/{chainID}/core/accounts/nftdata/{nftID} | Get the NFT data by an ID
[**accountsGetNativeTokenIDRegistry**](CorecontractsApi.md#accountsGetNativeTokenIDRegistry) | **GET** /v1/chains/{chainID}/core/accounts/token_registry | Get a list of all registries
[**accountsGetTotalAssets**](CorecontractsApi.md#accountsGetTotalAssets) | **GET** /v1/chains/{chainID}/core/accounts/total_assets | Get all stored assets
[**blobsGetBlobInfo**](CorecontractsApi.md#blobsGetBlobInfo) | **GET** /v1/chains/{chainID}/core/blobs/{blobHash} | Get all fields of a blob
[**blobsGetBlobValue**](CorecontractsApi.md#blobsGetBlobValue) | **GET** /v1/chains/{chainID}/core/blobs/{blobHash}/data/{fieldKey} | Get the value of the supplied field (key)
[**blocklogGetBlockInfo**](CorecontractsApi.md#blocklogGetBlockInfo) | **GET** /v1/chains/{chainID}/core/blocklog/blocks/{blockIndex} | Get the block info of a certain block index
[**blocklogGetControlAddresses**](CorecontractsApi.md#blocklogGetControlAddresses) | **GET** /v1/chains/{chainID}/core/blocklog/controladdresses | Get the control addresses
[**blocklogGetEventsOfBlock**](CorecontractsApi.md#blocklogGetEventsOfBlock) | **GET** /v1/chains/{chainID}/core/blocklog/events/block/{blockIndex} | Get events of a block
[**blocklogGetEventsOfLatestBlock**](CorecontractsApi.md#blocklogGetEventsOfLatestBlock) | **GET** /v1/chains/{chainID}/core/blocklog/events/block/latest | Get events of the latest block
[**blocklogGetEventsOfRequest**](CorecontractsApi.md#blocklogGetEventsOfRequest) | **GET** /v1/chains/{chainID}/core/blocklog/events/request/{requestID} | Get events of a request
[**blocklogGetLatestBlockInfo**](CorecontractsApi.md#blocklogGetLatestBlockInfo) | **GET** /v1/chains/{chainID}/core/blocklog/blocks/latest | Get the block info of the latest block
[**blocklogGetRequestIDsForBlock**](CorecontractsApi.md#blocklogGetRequestIDsForBlock) | **GET** /v1/chains/{chainID}/core/blocklog/blocks/{blockIndex}/requestids | Get the request ids for a certain block index
[**blocklogGetRequestIDsForLatestBlock**](CorecontractsApi.md#blocklogGetRequestIDsForLatestBlock) | **GET** /v1/chains/{chainID}/core/blocklog/blocks/latest/requestids | Get the request ids for the latest block
[**blocklogGetRequestIsProcessed**](CorecontractsApi.md#blocklogGetRequestIsProcessed) | **GET** /v1/chains/{chainID}/core/blocklog/requests/{requestID}/is_processed | Get the request processing status
[**blocklogGetRequestReceipt**](CorecontractsApi.md#blocklogGetRequestReceipt) | **GET** /v1/chains/{chainID}/core/blocklog/requests/{requestID} | Get the receipt of a certain request id
[**blocklogGetRequestReceiptsOfBlock**](CorecontractsApi.md#blocklogGetRequestReceiptsOfBlock) | **GET** /v1/chains/{chainID}/core/blocklog/blocks/{blockIndex}/receipts | Get all receipts of a certain block
[**blocklogGetRequestReceiptsOfLatestBlock**](CorecontractsApi.md#blocklogGetRequestReceiptsOfLatestBlock) | **GET** /v1/chains/{chainID}/core/blocklog/blocks/latest/receipts | Get all receipts of the latest block
[**errorsGetErrorMessageFormat**](CorecontractsApi.md#errorsGetErrorMessageFormat) | **GET** /v1/chains/{chainID}/core/errors/{contractHname}/message/{errorID} | Get the error message format of a specific error id
[**governanceGetAllowedStateControllerAddresses**](CorecontractsApi.md#governanceGetAllowedStateControllerAddresses) | **GET** /v1/chains/{chainID}/core/governance/allowedstatecontrollers | Get the allowed state controller addresses
[**governanceGetChainInfo**](CorecontractsApi.md#governanceGetChainInfo) | **GET** /v1/chains/{chainID}/core/governance/chaininfo | Get the chain info
[**governanceGetChainOwner**](CorecontractsApi.md#governanceGetChainOwner) | **GET** /v1/chains/{chainID}/core/governance/chainowner | Get the chain owner


# **accountsGetAccountBalance**
> AssetsResponse accountsGetAccountBalance()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetAccountBalanceRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | AgentID (Bech32 for WasmVM | Hex for EVM)
  agentID: "agentID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetAccountBalance(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **agentID** | [**string**] | AgentID (Bech32 for WasmVM | Hex for EVM) | defaults to undefined
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

# **accountsGetAccountFoundries**
> AccountFoundriesResponse accountsGetAccountFoundries()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetAccountFoundriesRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | AgentID (Bech32 for WasmVM | Hex for EVM)
  agentID: "agentID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetAccountFoundries(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **agentID** | [**string**] | AgentID (Bech32 for WasmVM | Hex for EVM) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**AccountFoundriesResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | All foundries owned by an account |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **accountsGetAccountNFTIDs**
> AccountNFTsResponse accountsGetAccountNFTIDs()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetAccountNFTIDsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | AgentID (Bech32 for WasmVM | Hex for EVM)
  agentID: "agentID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetAccountNFTIDs(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **agentID** | [**string**] | AgentID (Bech32 for WasmVM | Hex for EVM) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**AccountNFTsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | All NFT ids belonging to an account |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **accountsGetAccountNonce**
> AccountNonceResponse accountsGetAccountNonce()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetAccountNonceRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | AgentID (Bech32 for WasmVM | Hex for EVM)
  agentID: "agentID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetAccountNonce(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **agentID** | [**string**] | AgentID (Bech32 for WasmVM | Hex for EVM) | defaults to undefined
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

# **accountsGetFoundryOutput**
> FoundryOutputResponse accountsGetFoundryOutput()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetFoundryOutputRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // number | Serial Number (uint32)
  serialNumber: 1,
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetFoundryOutput(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **serialNumber** | [**number**] | Serial Number (uint32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**FoundryOutputResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The foundry output |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **accountsGetNFTData**
> NFTJSON accountsGetNFTData()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetNFTDataRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | NFT ID (Hex)
  nftID: "nftID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetNFTData(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **nftID** | [**string**] | NFT ID (Hex) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**NFTJSON**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The NFT data |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **accountsGetNativeTokenIDRegistry**
> NativeTokenIDRegistryResponse accountsGetNativeTokenIDRegistry()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetNativeTokenIDRegistryRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetNativeTokenIDRegistry(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**NativeTokenIDRegistryResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all registries |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **accountsGetTotalAssets**
> AssetsResponse accountsGetTotalAssets()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiAccountsGetTotalAssetsRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.accountsGetTotalAssets(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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

# **blobsGetBlobInfo**
> BlobInfoResponse blobsGetBlobInfo()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlobsGetBlobInfoRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | BlobHash (Hex)
  blobHash: "blobHash_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blobsGetBlobInfo(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **blobHash** | [**string**] | BlobHash (Hex) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**BlobInfoResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | All blob fields and their values |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blobsGetBlobValue**
> BlobValueResponse blobsGetBlobValue()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlobsGetBlobValueRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | BlobHash (Hex)
  blobHash: "blobHash_example",
  // string | FieldKey (String)
  fieldKey: "fieldKey_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blobsGetBlobValue(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **blobHash** | [**string**] | BlobHash (Hex) | defaults to undefined
 **fieldKey** | [**string**] | FieldKey (String) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**BlobValueResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The value of the supplied field (key) |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **blocklogGetBlockInfo**
> BlockInfoResponse blocklogGetBlockInfo()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetBlockInfoRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // number | BlockIndex (uint32)
  blockIndex: 1,
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetBlockInfo(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetControlAddressesRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetControlAddresses(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetEventsOfBlockRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // number | BlockIndex (uint32)
  blockIndex: 1,
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetEventsOfBlock(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetEventsOfLatestBlockRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetEventsOfLatestBlock(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetEventsOfRequestRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | RequestID (Hex)
  requestID: "requestID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetEventsOfRequest(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetLatestBlockInfoRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetLatestBlockInfo(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetRequestIDsForBlockRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // number | BlockIndex (uint32)
  blockIndex: 1,
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetRequestIDsForBlock(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetRequestIDsForLatestBlockRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetRequestIDsForLatestBlock(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetRequestIsProcessedRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | RequestID (Hex)
  requestID: "requestID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetRequestIsProcessed(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetRequestReceiptRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | RequestID (Hex)
  requestID: "requestID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetRequestReceipt(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetRequestReceiptsOfBlockRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // number | BlockIndex (uint32)
  blockIndex: 1,
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetRequestReceiptsOfBlock(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiBlocklogGetRequestReceiptsOfLatestBlockRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.blocklogGetRequestReceiptsOfLatestBlock(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiErrorsGetErrorMessageFormatRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Contract (Hname as Hex)
  contractHname: "contractHname_example",
  // number | Error Id (uint16)
  errorID: 1,
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.errorsGetErrorMessageFormat(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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

# **governanceGetAllowedStateControllerAddresses**
> GovAllowedStateControllerAddressesResponse governanceGetAllowedStateControllerAddresses()

Returns the allowed state controller addresses

### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiGovernanceGetAllowedStateControllerAddressesRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.governanceGetAllowedStateControllerAddresses(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**GovAllowedStateControllerAddressesResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The state controller addresses |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **governanceGetChainInfo**
> GovChainInfoResponse governanceGetChainInfo()

If you are using the common API functions, you most likely rather want to use '/v1/chains/:chainID' to get information about a chain.

### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiGovernanceGetChainInfoRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.governanceGetChainInfo(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
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

# **governanceGetChainOwner**
> GovChainOwnerResponse governanceGetChainOwner()

Returns the chain owner

### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .CorecontractsApi(configuration);

let body:.CorecontractsApiGovernanceGetChainOwnerRequest = {
  // string | ChainID (Bech32)
  chainID: "chainID_example",
  // string | Block index or trie root (optional)
  block: "block_example",
};

apiInstance.governanceGetChainOwner(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chainID** | [**string**] | ChainID (Bech32) | defaults to undefined
 **block** | [**string**] | Block index or trie root | (optional) defaults to undefined


### Return type

**GovChainOwnerResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The chain owner |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


