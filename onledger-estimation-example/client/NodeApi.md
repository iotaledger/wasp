# .NodeApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**distrustPeer**](NodeApi.md#distrustPeer) | **DELETE** /v1/node/peers/trusted/{peer} | Distrust a peering node
[**generateDKS**](NodeApi.md#generateDKS) | **POST** /v1/node/dks | Generate a new distributed key
[**getAllPeers**](NodeApi.md#getAllPeers) | **GET** /v1/node/peers | Get basic information about all configured peers
[**getConfiguration**](NodeApi.md#getConfiguration) | **GET** /v1/node/config | Return the Wasp configuration
[**getDKSInfo**](NodeApi.md#getDKSInfo) | **GET** /v1/node/dks/{sharedAddress} | Get information about the shared address DKS configuration
[**getInfo**](NodeApi.md#getInfo) | **GET** /v1/node/info | Returns private information about this node.
[**getPeeringIdentity**](NodeApi.md#getPeeringIdentity) | **GET** /v1/node/peers/identity | Get basic peer info of the current node
[**getTrustedPeers**](NodeApi.md#getTrustedPeers) | **GET** /v1/node/peers/trusted | Get trusted peers
[**getVersion**](NodeApi.md#getVersion) | **GET** /v1/node/version | Returns the node version.
[**ownerCertificate**](NodeApi.md#ownerCertificate) | **GET** /v1/node/owner/certificate | Gets the node owner
[**shutdownNode**](NodeApi.md#shutdownNode) | **POST** /v1/node/shutdown | Shut down the node
[**trustPeer**](NodeApi.md#trustPeer) | **POST** /v1/node/peers/trusted | Trust a peering node


# **distrustPeer**
> void distrustPeer()


### Example


```typescript
import { createConfiguration, NodeApi } from '';
import type { NodeApiDistrustPeerRequest } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request: NodeApiDistrustPeerRequest = {
    // Name or PubKey (hex) of the trusted peer
  peer: "peer_example",
};

const data = await apiInstance.distrustPeer(request);
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
**200** | Peer was successfully distrusted |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | Peer not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **generateDKS**
> DKSharesInfo generateDKS(dKSharesPostRequest)


### Example


```typescript
import { createConfiguration, NodeApi } from '';
import type { NodeApiGenerateDKSRequest } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request: NodeApiGenerateDKSRequest = {
    // Request parameters
  dKSharesPostRequest: {
    peerIdentities: [
      "peerIdentities_example",
    ],
    threshold: 1,
    timeoutMS: 1,
  },
};

const data = await apiInstance.generateDKS(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dKSharesPostRequest** | **DKSharesPostRequest**| Request parameters |


### Return type

**DKSharesInfo**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | DK shares info |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getAllPeers**
> Array<PeeringNodeStatusResponse> getAllPeers()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.getAllPeers(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**Array<PeeringNodeStatusResponse>**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all peers |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getConfiguration**
> { [key: string]: string; } getConfiguration()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.getConfiguration(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**{ [key: string]: string; }**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Dumped configuration |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getDKSInfo**
> DKSharesInfo getDKSInfo()


### Example


```typescript
import { createConfiguration, NodeApi } from '';
import type { NodeApiGetDKSInfoRequest } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request: NodeApiGetDKSInfoRequest = {
    // SharedAddress (Hex Address)
  sharedAddress: "sharedAddress_example",
};

const data = await apiInstance.getDKSInfo(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **sharedAddress** | [**string**] | SharedAddress (Hex Address) | defaults to undefined


### Return type

**DKSharesInfo**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | DK shares info |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | Shared address not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getInfo**
> InfoResponse getInfo()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.getInfo(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**InfoResponse**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Returns information about this node. |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getPeeringIdentity**
> PeeringNodeIdentityResponse getPeeringIdentity()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.getPeeringIdentity(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**PeeringNodeIdentityResponse**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | This node peering identity |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getTrustedPeers**
> Array<PeeringNodeIdentityResponse> getTrustedPeers()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.getTrustedPeers(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**Array<PeeringNodeIdentityResponse>**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of trusted peers |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getVersion**
> VersionResponse getVersion()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.getVersion(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**VersionResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Returns the version of the node. |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **ownerCertificate**
> NodeOwnerCertificateResponse ownerCertificate()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.ownerCertificate(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters
This endpoint does not need any parameter.


### Return type

**NodeOwnerCertificateResponse**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Node Certificate |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **shutdownNode**
> void shutdownNode()


### Example


```typescript
import { createConfiguration, NodeApi } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request = {};

const data = await apiInstance.shutdownNode(request);
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
**200** | The node has been shut down |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **trustPeer**
> void trustPeer(peeringTrustRequest)


### Example


```typescript
import { createConfiguration, NodeApi } from '';
import type { NodeApiTrustPeerRequest } from '';

const configuration = createConfiguration();
const apiInstance = new NodeApi(configuration);

const request: NodeApiTrustPeerRequest = {
    // Info of the peer to trust
  peeringTrustRequest: {
    name: "name_example",
    peeringURL: "peeringURL_example",
    publicKey: "publicKey_example",
  },
};

const data = await apiInstance.trustPeer(request);
console.log('API called successfully. Returned data:', data);
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **peeringTrustRequest** | **PeeringTrustRequest**| Info of the peer to trust |


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
**200** | Peer was successfully trusted |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


