# \NodeAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DistrustPeer**](NodeAPI.md#DistrustPeer) | **Delete** /v1/node/peers/trusted/{peer} | Distrust a peering node
[**GenerateDKS**](NodeAPI.md#GenerateDKS) | **Post** /v1/node/dks | Generate a new distributed key
[**GetAllPeers**](NodeAPI.md#GetAllPeers) | **Get** /v1/node/peers | Get basic information about all configured peers
[**GetConfiguration**](NodeAPI.md#GetConfiguration) | **Get** /v1/node/config | Return the Wasp configuration
[**GetDKSInfo**](NodeAPI.md#GetDKSInfo) | **Get** /v1/node/dks/{sharedAddress} | Get information about the shared address DKS configuration
[**GetInfo**](NodeAPI.md#GetInfo) | **Get** /v1/node/info | Returns private information about this node.
[**GetPeeringIdentity**](NodeAPI.md#GetPeeringIdentity) | **Get** /v1/node/peers/identity | Get basic peer info of the current node
[**GetTrustedPeers**](NodeAPI.md#GetTrustedPeers) | **Get** /v1/node/peers/trusted | Get trusted peers
[**GetVersion**](NodeAPI.md#GetVersion) | **Get** /v1/node/version | Returns the node version.
[**OwnerCertificate**](NodeAPI.md#OwnerCertificate) | **Get** /v1/node/owner/certificate | Gets the node owner
[**ShutdownNode**](NodeAPI.md#ShutdownNode) | **Post** /v1/node/shutdown | Shut down the node
[**TrustPeer**](NodeAPI.md#TrustPeer) | **Post** /v1/node/peers/trusted | Trust a peering node



## DistrustPeer

> DistrustPeer(ctx, peer).Execute()

Distrust a peering node

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	peer := "peer_example" // string | Name or PubKey (hex) of the trusted peer

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodeAPI.DistrustPeer(context.Background(), peer).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.DistrustPeer``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**peer** | **string** | Name or PubKey (hex) of the trusted peer | 

### Other Parameters

Other parameters are passed through a pointer to a apiDistrustPeerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## GenerateDKS

> DKSharesInfo GenerateDKS(ctx).DKSharesPostRequest(dKSharesPostRequest).Execute()

Generate a new distributed key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	dKSharesPostRequest := *openapiclient.NewDKSharesPostRequest([]string{"PeerIdentities_example"}, uint32(123), uint32(123)) // DKSharesPostRequest | Request parameters

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GenerateDKS(context.Background()).DKSharesPostRequest(dKSharesPostRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GenerateDKS``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GenerateDKS`: DKSharesInfo
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GenerateDKS`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGenerateDKSRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dKSharesPostRequest** | [**DKSharesPostRequest**](DKSharesPostRequest.md) | Request parameters | 

### Return type

[**DKSharesInfo**](DKSharesInfo.md)

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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GetAllPeers(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GetAllPeers``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAllPeers`: []PeeringNodeStatusResponse
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GetAllPeers`: %v\n", resp)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GetConfiguration(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GetConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetConfiguration`: map[string]string
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GetConfiguration`: %v\n", resp)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	sharedAddress := "sharedAddress_example" // string | SharedAddress (Hex Address)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GetDKSInfo(context.Background(), sharedAddress).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GetDKSInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDKSInfo`: DKSharesInfo
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GetDKSInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sharedAddress** | **string** | SharedAddress (Hex Address) | 

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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GetInfo(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GetInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInfo`: InfoResponse
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GetInfo`: %v\n", resp)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GetPeeringIdentity(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GetPeeringIdentity``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPeeringIdentity`: PeeringNodeIdentityResponse
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GetPeeringIdentity`: %v\n", resp)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GetTrustedPeers(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GetTrustedPeers``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTrustedPeers`: []PeeringNodeIdentityResponse
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GetTrustedPeers`: %v\n", resp)
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

> VersionResponse GetVersion(ctx).Execute()

Returns the node version.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.GetVersion(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.GetVersion``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVersion`: VersionResponse
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.GetVersion`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetVersionRequest struct via the builder pattern


### Return type

[**VersionResponse**](VersionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OwnerCertificate

> NodeOwnerCertificateResponse OwnerCertificate(ctx).Execute()

Gets the node owner

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeAPI.OwnerCertificate(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.OwnerCertificate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OwnerCertificate`: NodeOwnerCertificateResponse
	fmt.Fprintf(os.Stdout, "Response from `NodeAPI.OwnerCertificate`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiOwnerCertificateRequest struct via the builder pattern


### Return type

[**NodeOwnerCertificateResponse**](NodeOwnerCertificateResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodeAPI.ShutdownNode(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.ShutdownNode``: %v\n", err)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	peeringTrustRequest := *openapiclient.NewPeeringTrustRequest("Name_example", "PeeringURL_example", "PublicKey_example") // PeeringTrustRequest | Info of the peer to trust

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodeAPI.TrustPeer(context.Background()).PeeringTrustRequest(peeringTrustRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeAPI.TrustPeer``: %v\n", err)
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

