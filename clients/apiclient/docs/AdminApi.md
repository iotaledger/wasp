# \AdminApi

All URIs are relative to *http://localhost:9090*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AdmChainChainIDAccessNodePubKeyDelete**](AdminApi.md#AdmChainChainIDAccessNodePubKeyDelete) | **Delete** /adm/chain/{chainID}/access-node/{pubKey} | Remove an access node from a chain
[**AdmChainChainIDAccessNodePubKeyPut**](AdminApi.md#AdmChainChainIDAccessNodePubKeyPut) | **Put** /adm/chain/{chainID}/access-node/{pubKey} | Add an access node to a chain
[**AdmChainChainIDActivatePost**](AdminApi.md#AdmChainChainIDActivatePost) | **Post** /adm/chain/{chainID}/activate | Activate a chain
[**AdmChainChainIDConsensusMetricsPipeGet**](AdminApi.md#AdmChainChainIDConsensusMetricsPipeGet) | **Get** /adm/chain/{chainID}/consensus/metrics/pipe | Get consensus pipe metrics
[**AdmChainChainIDConsensusStatusGet**](AdminApi.md#AdmChainChainIDConsensusStatusGet) | **Get** /adm/chain/{chainID}/consensus/status | Get chain state statistics for the given chain ID
[**AdmChainChainIDDeactivatePost**](AdminApi.md#AdmChainChainIDDeactivatePost) | **Post** /adm/chain/{chainID}/deactivate | Deactivate a chain
[**AdmChainChainIDInfoGet**](AdminApi.md#AdmChainChainIDInfoGet) | **Get** /adm/chain/{chainID}/info | Get basic chain info.
[**AdmChainChainIDNodeconnMetricsGet**](AdminApi.md#AdmChainChainIDNodeconnMetricsGet) | **Get** /adm/chain/{chainID}/nodeconn/metrics | Get chain node connection metrics for the given chain ID
[**AdmChainNodeconnMetricsGet**](AdminApi.md#AdmChainNodeconnMetricsGet) | **Get** /adm/chain/nodeconn/metrics | Get cummulative chains node connection metrics
[**AdmChainrecordChainIDGet**](AdminApi.md#AdmChainrecordChainIDGet) | **Get** /adm/chainrecord/{chainID} | Find the chain record for the given chain ID
[**AdmChainrecordPost**](AdminApi.md#AdmChainrecordPost) | **Post** /adm/chainrecord | Create a new chain record
[**AdmChainrecordsGet**](AdminApi.md#AdmChainrecordsGet) | **Get** /adm/chainrecords | Get the list of chain records in the node
[**AdmDksPost**](AdminApi.md#AdmDksPost) | **Post** /adm/dks | Generate a new distributed key
[**AdmDksSharedAddressGet**](AdminApi.md#AdmDksSharedAddressGet) | **Get** /adm/dks/{sharedAddress} | Get distributed key properties
[**AdmNodeOwnerCertificatePost**](AdminApi.md#AdmNodeOwnerCertificatePost) | **Post** /adm/node/owner/certificate | Provides a certificate, if the node recognizes the owner.
[**AdmPeeringEstablishedGet**](AdminApi.md#AdmPeeringEstablishedGet) | **Get** /adm/peering/established | Basic information about all configured peers.
[**AdmPeeringSelfGet**](AdminApi.md#AdmPeeringSelfGet) | **Get** /adm/peering/self | Basic peer info of the current node.
[**AdmPeeringTrustedGet**](AdminApi.md#AdmPeeringTrustedGet) | **Get** /adm/peering/trusted | Get a list of trusted peers.
[**AdmPeeringTrustedPost**](AdminApi.md#AdmPeeringTrustedPost) | **Post** /adm/peering/trusted | Trust the specified peer.
[**AdmPeeringTrustedPubKeyDelete**](AdminApi.md#AdmPeeringTrustedPubKeyDelete) | **Delete** /adm/peering/trusted/{pubKey} | Distrust the specified peer.
[**AdmPeeringTrustedPubKeyGet**](AdminApi.md#AdmPeeringTrustedPubKeyGet) | **Get** /adm/peering/trusted/{pubKey} | Get details on a particular trusted peer.
[**AdmPeeringTrustedPubKeyPut**](AdminApi.md#AdmPeeringTrustedPubKeyPut) | **Put** /adm/peering/trusted/{pubKey} | Trust the specified peer, the pub key is passed via the path.
[**AdmShutdownGet**](AdminApi.md#AdmShutdownGet) | **Get** /adm/shutdown | Shut down the node



## AdmChainChainIDAccessNodePubKeyDelete

> AdmChainChainIDAccessNodePubKeyDelete(ctx, chainID, pubKey).Execute()

Remove an access node from a chain

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
    chainID := "chainID_example" // string | ChainID (bech32))
    pubKey := "pubKey_example" // string | PublicKey (hex string)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDAccessNodePubKeyDelete(context.Background(), chainID, pubKey).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDAccessNodePubKeyDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32)) | 
**pubKey** | **string** | PublicKey (hex string) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDAccessNodePubKeyDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainChainIDAccessNodePubKeyPut

> AdmChainChainIDAccessNodePubKeyPut(ctx, chainID, pubKey).Execute()

Add an access node to a chain

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
    chainID := "chainID_example" // string | ChainID (bech32))
    pubKey := "pubKey_example" // string | PublicKey (hex string)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDAccessNodePubKeyPut(context.Background(), chainID, pubKey).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDAccessNodePubKeyPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32)) | 
**pubKey** | **string** | PublicKey (hex string) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDAccessNodePubKeyPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainChainIDActivatePost

> AdmChainChainIDActivatePost(ctx, chainID).Execute()

Activate a chain

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
    chainID := "chainID_example" // string | ChainID (bech32))

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDActivatePost(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDActivatePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32)) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDActivatePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainChainIDConsensusMetricsPipeGet

> ConsensusPipeMetrics AdmChainChainIDConsensusMetricsPipeGet(ctx, chainID).Execute()

Get consensus pipe metrics

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
    chainID := "chainID_example" // string | chainID

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDConsensusMetricsPipeGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDConsensusMetricsPipeGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmChainChainIDConsensusMetricsPipeGet`: ConsensusPipeMetrics
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmChainChainIDConsensusMetricsPipeGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | chainID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDConsensusMetricsPipeGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ConsensusPipeMetrics**](ConsensusPipeMetrics.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainChainIDConsensusStatusGet

> ConsensusWorkflowStatus AdmChainChainIDConsensusStatusGet(ctx, chainID).Execute()

Get chain state statistics for the given chain ID

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
    chainID := "chainID_example" // string | ChainID (bech32)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDConsensusStatusGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDConsensusStatusGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmChainChainIDConsensusStatusGet`: ConsensusWorkflowStatus
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmChainChainIDConsensusStatusGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDConsensusStatusGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ConsensusWorkflowStatus**](ConsensusWorkflowStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainChainIDDeactivatePost

> AdmChainChainIDDeactivatePost(ctx, chainID).Execute()

Deactivate a chain

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
    chainID := "chainID_example" // string | ChainID (bech32))

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDDeactivatePost(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDDeactivatePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32)) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDDeactivatePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainChainIDInfoGet

> AdmChainChainIDInfoGet(ctx, chainID).Execute()

Get basic chain info.

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
    chainID := "chainID_example" // string | ChainID (bech32))

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDInfoGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDInfoGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32)) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDInfoGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainChainIDNodeconnMetricsGet

> NodeConnectionMessagesMetrics AdmChainChainIDNodeconnMetricsGet(ctx, chainID).Execute()

Get chain node connection metrics for the given chain ID

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
    chainID := "chainID_example" // string | ChainID (bech32)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainChainIDNodeconnMetricsGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainChainIDNodeconnMetricsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmChainChainIDNodeconnMetricsGet`: NodeConnectionMessagesMetrics
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmChainChainIDNodeconnMetricsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainChainIDNodeconnMetricsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**NodeConnectionMessagesMetrics**](NodeConnectionMessagesMetrics.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainNodeconnMetricsGet

> NodeConnectionMetrics AdmChainNodeconnMetricsGet(ctx).Execute()

Get cummulative chains node connection metrics

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
    resp, r, err := apiClient.AdminApi.AdmChainNodeconnMetricsGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainNodeconnMetricsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmChainNodeconnMetricsGet`: NodeConnectionMetrics
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmChainNodeconnMetricsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainNodeconnMetricsGetRequest struct via the builder pattern


### Return type

[**NodeConnectionMetrics**](NodeConnectionMetrics.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainrecordChainIDGet

> ChainRecord AdmChainrecordChainIDGet(ctx, chainID).Execute()

Find the chain record for the given chain ID

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
    chainID := "chainID_example" // string | ChainID (bech32)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainrecordChainIDGet(context.Background(), chainID).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainrecordChainIDGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmChainrecordChainIDGet`: ChainRecord
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmChainrecordChainIDGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (bech32) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainrecordChainIDGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ChainRecord**](ChainRecord.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainrecordPost

> AdmChainrecordPost(ctx).Record(record).Execute()

Create a new chain record

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
    record := *openapiclient.NewChainRecord() // ChainRecord | Chain record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmChainrecordPost(context.Background()).Record(record).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainrecordPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainrecordPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **record** | [**ChainRecord**](ChainRecord.md) | Chain record | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmChainrecordsGet

> []ChainRecord AdmChainrecordsGet(ctx).Execute()

Get the list of chain records in the node

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
    resp, r, err := apiClient.AdminApi.AdmChainrecordsGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmChainrecordsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmChainrecordsGet`: []ChainRecord
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmChainrecordsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAdmChainrecordsGetRequest struct via the builder pattern


### Return type

[**[]ChainRecord**](ChainRecord.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmDksPost

> DKSharesInfo AdmDksPost(ctx).DKSharesPostRequest(dKSharesPostRequest).Execute()

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
    resp, r, err := apiClient.AdminApi.AdmDksPost(context.Background()).DKSharesPostRequest(dKSharesPostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmDksPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmDksPost`: DKSharesInfo
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmDksPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAdmDksPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dKSharesPostRequest** | [**DKSharesPostRequest**](DKSharesPostRequest.md) | Request parameters | 

### Return type

[**DKSharesInfo**](DKSharesInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmDksSharedAddressGet

> DKSharesInfo AdmDksSharedAddressGet(ctx, sharedAddress).Execute()

Get distributed key properties

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
    sharedAddress := "sharedAddress_example" // string | Address of the DK share (hex)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmDksSharedAddressGet(context.Background(), sharedAddress).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmDksSharedAddressGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmDksSharedAddressGet`: DKSharesInfo
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmDksSharedAddressGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sharedAddress** | **string** | Address of the DK share (hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmDksSharedAddressGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DKSharesInfo**](DKSharesInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmNodeOwnerCertificatePost

> NodeOwnerCertificateResponse AdmNodeOwnerCertificatePost(ctx).Request(request).Execute()

Provides a certificate, if the node recognizes the owner.

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
    request := *openapiclient.NewNodeOwnerCertificateRequest("OwnerAddress_example", "PublicKey_example") // NodeOwnerCertificateRequest | Certificate request

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmNodeOwnerCertificatePost(context.Background()).Request(request).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmNodeOwnerCertificatePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmNodeOwnerCertificatePost`: NodeOwnerCertificateResponse
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmNodeOwnerCertificatePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAdmNodeOwnerCertificatePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**NodeOwnerCertificateRequest**](NodeOwnerCertificateRequest.md) | Certificate request | 

### Return type

[**NodeOwnerCertificateResponse**](NodeOwnerCertificateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmPeeringEstablishedGet

> []PeeringNodeStatus AdmPeeringEstablishedGet(ctx).Execute()

Basic information about all configured peers.

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
    resp, r, err := apiClient.AdminApi.AdmPeeringEstablishedGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmPeeringEstablishedGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmPeeringEstablishedGet`: []PeeringNodeStatus
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmPeeringEstablishedGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAdmPeeringEstablishedGetRequest struct via the builder pattern


### Return type

[**[]PeeringNodeStatus**](PeeringNodeStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmPeeringSelfGet

> PeeringTrustedNode AdmPeeringSelfGet(ctx).Execute()

Basic peer info of the current node.

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
    resp, r, err := apiClient.AdminApi.AdmPeeringSelfGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmPeeringSelfGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmPeeringSelfGet`: PeeringTrustedNode
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmPeeringSelfGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAdmPeeringSelfGetRequest struct via the builder pattern


### Return type

[**PeeringTrustedNode**](PeeringTrustedNode.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmPeeringTrustedGet

> []PeeringTrustedNode AdmPeeringTrustedGet(ctx).Execute()

Get a list of trusted peers.

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
    resp, r, err := apiClient.AdminApi.AdmPeeringTrustedGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmPeeringTrustedGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmPeeringTrustedGet`: []PeeringTrustedNode
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmPeeringTrustedGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAdmPeeringTrustedGetRequest struct via the builder pattern


### Return type

[**[]PeeringTrustedNode**](PeeringTrustedNode.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmPeeringTrustedPost

> PeeringTrustedNode AdmPeeringTrustedPost(ctx).PeeringTrustedNode(peeringTrustedNode).Execute()

Trust the specified peer.

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
    peeringTrustedNode := *openapiclient.NewPeeringTrustedNode() // PeeringTrustedNode | Info of the peer to trust.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmPeeringTrustedPost(context.Background()).PeeringTrustedNode(peeringTrustedNode).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmPeeringTrustedPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmPeeringTrustedPost`: PeeringTrustedNode
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmPeeringTrustedPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAdmPeeringTrustedPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **peeringTrustedNode** | [**PeeringTrustedNode**](PeeringTrustedNode.md) | Info of the peer to trust. | 

### Return type

[**PeeringTrustedNode**](PeeringTrustedNode.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmPeeringTrustedPubKeyDelete

> AdmPeeringTrustedPubKeyDelete(ctx, pubKey).Execute()

Distrust the specified peer.

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
    pubKey := "pubKey_example" // string | Public key of the trusted peer (hex).

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmPeeringTrustedPubKeyDelete(context.Background(), pubKey).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmPeeringTrustedPubKeyDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**pubKey** | **string** | Public key of the trusted peer (hex). | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmPeeringTrustedPubKeyDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmPeeringTrustedPubKeyGet

> PeeringTrustedNode AdmPeeringTrustedPubKeyGet(ctx, pubKey).Execute()

Get details on a particular trusted peer.

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
    pubKey := "pubKey_example" // string | Public key of the trusted peer (hex).

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmPeeringTrustedPubKeyGet(context.Background(), pubKey).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmPeeringTrustedPubKeyGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmPeeringTrustedPubKeyGet`: PeeringTrustedNode
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmPeeringTrustedPubKeyGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**pubKey** | **string** | Public key of the trusted peer (hex). | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmPeeringTrustedPubKeyGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**PeeringTrustedNode**](PeeringTrustedNode.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmPeeringTrustedPubKeyPut

> PeeringTrustedNode AdmPeeringTrustedPubKeyPut(ctx, pubKey).PeeringTrustedNode(peeringTrustedNode).Execute()

Trust the specified peer, the pub key is passed via the path.

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
    pubKey := "pubKey_example" // string | Public key of the trusted peer (hex).
    peeringTrustedNode := *openapiclient.NewPeeringTrustedNode() // PeeringTrustedNode | Info of the peer to trust.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdminApi.AdmPeeringTrustedPubKeyPut(context.Background(), pubKey).PeeringTrustedNode(peeringTrustedNode).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmPeeringTrustedPubKeyPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdmPeeringTrustedPubKeyPut`: PeeringTrustedNode
    fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdmPeeringTrustedPubKeyPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**pubKey** | **string** | Public key of the trusted peer (hex). | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdmPeeringTrustedPubKeyPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **peeringTrustedNode** | [**PeeringTrustedNode**](PeeringTrustedNode.md) | Info of the peer to trust. | 

### Return type

[**PeeringTrustedNode**](PeeringTrustedNode.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdmShutdownGet

> AdmShutdownGet(ctx).Execute()

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
    resp, r, err := apiClient.AdminApi.AdmShutdownGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdmShutdownGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAdmShutdownGetRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

