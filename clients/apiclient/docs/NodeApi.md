# \NodeApi

All URIs are relative to *http://localhost:9090*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DistrustPeer**](NodeApi.md#DistrustPeer) | **Delete** /v2node/peers/trusted | Distrust a peering node
[**GenerateDKS**](NodeApi.md#GenerateDKS) | **Post** /v2node/dks | Generate a new distributed key
[**GetAllPeers**](NodeApi.md#GetAllPeers) | **Get** /v2node/peers | Get basic information about all configured peers
[**GetConfiguration**](NodeApi.md#GetConfiguration) | **Get** /v2node/config | Return the Wasp configuration
[**GetDKSInfo**](NodeApi.md#GetDKSInfo) | **Get** /v2node/dks/{sharedAddress} | Get information about the shared address DKS configuration
[**GetInfo**](NodeApi.md#GetInfo) | **Get** /v2node/info | Returns private information about this node.
[**GetPeeringIdentity**](NodeApi.md#GetPeeringIdentity) | **Get** /v2node/peers/identity | Get basic peer info of the current node
[**GetTrustedPeers**](NodeApi.md#GetTrustedPeers) | **Get** /v2node/peers/trusted | Get trusted peers
[**GetVersion**](NodeApi.md#GetVersion) | **Get** /v2node/version | Returns the node version.
[**SetNodeOwner**](NodeApi.md#SetNodeOwner) | **Post** /v2node/owner/certificate | Sets the node owner
[**ShutdownNode**](NodeApi.md#ShutdownNode) | **Post** /v2node/shutdown | Shut down the node
[**TrustPeer**](NodeApi.md#TrustPeer) | **Post** /v2node/peers/trusted | Trust a peering node



## DistrustPeer

> DistrustPeer(ctx).PeeringTrustRequest(peeringTrustRequest).Execute()

Distrust a peering node

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    peeringTrustRequest := *openapiclient.NewPeeringTrustRequest("localhost:4000", "0x0000") // PeeringTrustRequest | Info of the peer to distrust

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.DistrustPeer(context.Background()).PeeringTrustRequest(peeringTrustRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.DistrustPeer``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDistrustPeerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **peeringTrustRequest** | [**PeeringTrustRequest**](PeeringTrustRequest.md) | Info of the peer to distrust | 

### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GenerateDKS

> DKSharesPostRequest GenerateDKS(ctx).DKSharesPostRequest(dKSharesPostRequest).Execute()

Generate a new distributed key

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    dKSharesPostRequest := *openapiclient.NewDKSharesPostRequest([]string{"PeerIdentities_example"}, uint32(123), uint32(123)) // DKSharesPostRequest | Request parameters

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GenerateDKS(context.Background()).DKSharesPostRequest(dKSharesPostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GenerateDKS``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GenerateDKS`: DKSharesPostRequest
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GenerateDKS`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGenerateDKSRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dKSharesPostRequest** | [**DKSharesPostRequest**](DKSharesPostRequest.md) | Request parameters | 

### Return type

[**DKSharesPostRequest**](DKSharesPostRequest.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAllPeers

> []PeeringNodeStatusResponse GetAllPeers(ctx).Execute()

Get basic information about all configured peers

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GetAllPeers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GetAllPeers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAllPeers`: []PeeringNodeStatusResponse
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GetAllPeers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetAllPeersRequest struct via the builder pattern


### Return type

[**[]PeeringNodeStatusResponse**](PeeringNodeStatusResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetConfiguration

> map[string]string GetConfiguration(ctx).Execute()

Return the Wasp configuration

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GetConfiguration(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GetConfiguration``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetConfiguration`: map[string]string
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GetConfiguration`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetConfigurationRequest struct via the builder pattern


### Return type

**map[string]string**

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDKSInfo

> DKSharesInfo GetDKSInfo(ctx, sharedAddress).Execute()

Get information about the shared address DKS configuration

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    sharedAddress := "sharedAddress_example" // string | SharedAddress (Bech32)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GetDKSInfo(context.Background(), sharedAddress).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GetDKSInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDKSInfo`: DKSharesInfo
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GetDKSInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sharedAddress** | **string** | SharedAddress (Bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDKSInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DKSharesInfo**](DKSharesInfo.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInfo

> InfoResponse GetInfo(ctx).Execute()

Returns private information about this node.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GetInfo(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GetInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInfo`: InfoResponse
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GetInfo`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInfoRequest struct via the builder pattern


### Return type

[**InfoResponse**](InfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPeeringIdentity

> PeeringNodeIdentityResponse GetPeeringIdentity(ctx).Execute()

Get basic peer info of the current node

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GetPeeringIdentity(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GetPeeringIdentity``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPeeringIdentity`: PeeringNodeIdentityResponse
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GetPeeringIdentity`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetPeeringIdentityRequest struct via the builder pattern


### Return type

[**PeeringNodeIdentityResponse**](PeeringNodeIdentityResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTrustedPeers

> []PeeringNodeIdentityResponse GetTrustedPeers(ctx).Execute()

Get trusted peers

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GetTrustedPeers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GetTrustedPeers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTrustedPeers`: []PeeringNodeIdentityResponse
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GetTrustedPeers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetTrustedPeersRequest struct via the builder pattern


### Return type

[**[]PeeringNodeIdentityResponse**](PeeringNodeIdentityResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetVersion

> string GetVersion(ctx).Execute()

Returns the node version.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.GetVersion(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.GetVersion``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetVersion`: string
    fmt.Fprintf(os.Stdout, "Response from `NodeApi.GetVersion`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetVersionRequest struct via the builder pattern


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetNodeOwner

> SetNodeOwner(ctx).NodeOwnerCertificateRequest(nodeOwnerCertificateRequest).Execute()

Sets the node owner

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    nodeOwnerCertificateRequest := *openapiclient.NewNodeOwnerCertificateRequest("OwnerAddress_example", "PublicKey_example") // NodeOwnerCertificateRequest | The node owner certificate

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.SetNodeOwner(context.Background()).NodeOwnerCertificateRequest(nodeOwnerCertificateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.SetNodeOwner``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetNodeOwnerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nodeOwnerCertificateRequest** | [**NodeOwnerCertificateRequest**](NodeOwnerCertificateRequest.md) | The node owner certificate | 

### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ShutdownNode

> ShutdownNode(ctx).Execute()

Shut down the node

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.ShutdownNode(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.ShutdownNode``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiShutdownNodeRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## TrustPeer

> TrustPeer(ctx).PeeringTrustRequest(peeringTrustRequest).Execute()

Trust a peering node

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    peeringTrustRequest := *openapiclient.NewPeeringTrustRequest("localhost:4000", "0x0000") // PeeringTrustRequest | Info of the peer to trust

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodeApi.TrustPeer(context.Background()).PeeringTrustRequest(peeringTrustRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodeApi.TrustPeer``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiTrustPeerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **peeringTrustRequest** | [**PeeringTrustRequest**](PeeringTrustRequest.md) | Info of the peer to trust | 

### Return type

 (empty response body)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

