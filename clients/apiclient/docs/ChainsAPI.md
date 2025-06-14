# \ChainsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ActivateChain**](ChainsAPI.md#ActivateChain) | **Post** /v1/chain/activate/{chainID} | Activate a chain
[**AddAccessNode**](ChainsAPI.md#AddAccessNode) | **Put** /v1/chain/access-node/{peer} | Configure a trusted node to be an access node.
[**CallView**](ChainsAPI.md#CallView) | **Post** /v1/chain/callview | Call a view function on a contract by Hname
[**DeactivateChain**](ChainsAPI.md#DeactivateChain) | **Post** /v1/chain/deactivate | Deactivate a chain
[**DumpAccounts**](ChainsAPI.md#DumpAccounts) | **Post** /v1/chain/dump-accounts | dump accounts information into a humanly-readable format
[**EstimateGasOffledger**](ChainsAPI.md#EstimateGasOffledger) | **Post** /v1/chain/estimategas-offledger | Estimates gas for a given off-ledger ISC request
[**EstimateGasOnledger**](ChainsAPI.md#EstimateGasOnledger) | **Post** /v1/chain/estimategas-onledger | Estimates gas for a given on-ledger ISC request
[**GetChainInfo**](ChainsAPI.md#GetChainInfo) | **Get** /v1/chain | Get information about a specific chain
[**GetCommitteeInfo**](ChainsAPI.md#GetCommitteeInfo) | **Get** /v1/chain/committee | Get information about the deployed committee
[**GetContracts**](ChainsAPI.md#GetContracts) | **Get** /v1/chain/contracts | Get all available chain contracts
[**GetMempoolContents**](ChainsAPI.md#GetMempoolContents) | **Get** /v1/chain/mempool | Get the contents of the mempool.
[**GetReceipt**](ChainsAPI.md#GetReceipt) | **Get** /v1/chain/receipts/{requestID} | Get a receipt from a request ID
[**GetStateValue**](ChainsAPI.md#GetStateValue) | **Get** /v1/chain/state/{stateKey} | Fetch the raw value associated with the given key in the chain state
[**RemoveAccessNode**](ChainsAPI.md#RemoveAccessNode) | **Delete** /v1/chain/access-node/{peer} | Remove an access node.
[**RotateChain**](ChainsAPI.md#RotateChain) | **Post** /v1/chain/rotate | Rotate a chain
[**SetChainRecord**](ChainsAPI.md#SetChainRecord) | **Post** /v1/chain/chainrecord/{chainID} | Sets the chain record.
[**V1ChainEvmPost**](ChainsAPI.md#V1ChainEvmPost) | **Post** /v1/chain/evm | Ethereum JSON-RPC
[**V1ChainEvmWsGet**](ChainsAPI.md#V1ChainEvmWsGet) | **Get** /v1/chain/evm/ws | Ethereum JSON-RPC (Websocket transport)
[**WaitForRequest**](ChainsAPI.md#WaitForRequest) | **Get** /v1/chain/requests/{requestID}/wait | Wait until the given request has been processed by the node



## ActivateChain

> ActivateChain(ctx, chainID).Execute()

Activate a chain

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
	chainID := "chainID_example" // string | ChainID (Hex Address)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ChainsAPI.ActivateChain(context.Background(), chainID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.ActivateChain``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Hex Address) | 

### Other Parameters

Other parameters are passed through a pointer to a apiActivateChainRequest struct via the builder pattern


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


## AddAccessNode

> AddAccessNode(ctx, peer).Execute()

Configure a trusted node to be an access node.

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
	r, err := apiClient.ChainsAPI.AddAccessNode(context.Background(), peer).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.AddAccessNode``: %v\n", err)
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

Other parameters are passed through a pointer to a apiAddAccessNodeRequest struct via the builder pattern


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


## CallView

> []string CallView(ctx).ContractCallViewRequest(contractCallViewRequest).Execute()

Call a view function on a contract by Hname



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
	contractCallViewRequest := *openapiclient.NewContractCallViewRequest([]string{"Arguments_example"}, "ContractHName_example", "ContractName_example", "FunctionHName_example", "FunctionName_example") // ContractCallViewRequest | Parameters

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.CallView(context.Background()).ContractCallViewRequest(contractCallViewRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.CallView``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CallView`: []string
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.CallView`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCallViewRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **contractCallViewRequest** | [**ContractCallViewRequest**](ContractCallViewRequest.md) | Parameters | 

### Return type

**[]string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeactivateChain

> DeactivateChain(ctx).Execute()

Deactivate a chain

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
	r, err := apiClient.ChainsAPI.DeactivateChain(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.DeactivateChain``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeactivateChainRequest struct via the builder pattern


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


## DumpAccounts

> DumpAccounts(ctx).Execute()

dump accounts information into a humanly-readable format

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
	r, err := apiClient.ChainsAPI.DumpAccounts(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.DumpAccounts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDumpAccountsRequest struct via the builder pattern


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


## EstimateGasOffledger

> ReceiptResponse EstimateGasOffledger(ctx).Request(request).Execute()

Estimates gas for a given off-ledger ISC request

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
	request := *openapiclient.NewEstimateGasRequestOffledger("FromAddress_example", "RequestBytes_example") // EstimateGasRequestOffledger | Request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.EstimateGasOffledger(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.EstimateGasOffledger``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `EstimateGasOffledger`: ReceiptResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.EstimateGasOffledger`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiEstimateGasOffledgerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**EstimateGasRequestOffledger**](EstimateGasRequestOffledger.md) | Request | 

### Return type

[**ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## EstimateGasOnledger

> OnLedgerEstimationResponse EstimateGasOnledger(ctx).Request(request).Execute()

Estimates gas for a given on-ledger ISC request

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
	request := *openapiclient.NewEstimateGasRequestOnledger("TransactionBytes_example") // EstimateGasRequestOnledger | Request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.EstimateGasOnledger(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.EstimateGasOnledger``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `EstimateGasOnledger`: OnLedgerEstimationResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.EstimateGasOnledger`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiEstimateGasOnledgerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**EstimateGasRequestOnledger**](EstimateGasRequestOnledger.md) | Request | 

### Return type

[**OnLedgerEstimationResponse**](OnLedgerEstimationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetChainInfo

> ChainInfoResponse GetChainInfo(ctx).Block(block).Execute()

Get information about a specific chain

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
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.GetChainInfo(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.GetChainInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetChainInfo`: ChainInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.GetChainInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetChainInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**ChainInfoResponse**](ChainInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCommitteeInfo

> CommitteeInfoResponse GetCommitteeInfo(ctx).Block(block).Execute()

Get information about the deployed committee

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
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.GetCommitteeInfo(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.GetCommitteeInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetCommitteeInfo`: CommitteeInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.GetCommitteeInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetCommitteeInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**CommitteeInfoResponse**](CommitteeInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetContracts

> []ContractInfoResponse GetContracts(ctx).Block(block).Execute()

Get all available chain contracts

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
	block := "block_example" // string | Block index or trie root (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.GetContracts(context.Background()).Block(block).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.GetContracts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetContracts`: []ContractInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.GetContracts`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetContractsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **block** | **string** | Block index or trie root | 

### Return type

[**[]ContractInfoResponse**](ContractInfoResponse.md)

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMempoolContents

> []int32 GetMempoolContents(ctx).Execute()

Get the contents of the mempool.

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
	resp, r, err := apiClient.ChainsAPI.GetMempoolContents(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.GetMempoolContents``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMempoolContents`: []int32
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.GetMempoolContents`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetMempoolContentsRequest struct via the builder pattern


### Return type

**[]int32**

### Authorization

[Authorization](../README.md#Authorization)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetReceipt

> ReceiptResponse GetReceipt(ctx, requestID).Execute()

Get a receipt from a request ID

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
	requestID := "requestID_example" // string | RequestID (Hex)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.GetReceipt(context.Background(), requestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.GetReceipt``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetReceipt`: ReceiptResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.GetReceipt`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetReceiptRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetStateValue

> StateResponse GetStateValue(ctx, stateKey).Execute()

Fetch the raw value associated with the given key in the chain state

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
	stateKey := "stateKey_example" // string | State Key (Hex)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.GetStateValue(context.Background(), stateKey).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.GetStateValue``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetStateValue`: StateResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.GetStateValue`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**stateKey** | **string** | State Key (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetStateValueRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**StateResponse**](StateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveAccessNode

> RemoveAccessNode(ctx, peer).Execute()

Remove an access node.

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
	r, err := apiClient.ChainsAPI.RemoveAccessNode(context.Background(), peer).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.RemoveAccessNode``: %v\n", err)
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

Other parameters are passed through a pointer to a apiRemoveAccessNodeRequest struct via the builder pattern


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


## RotateChain

> RotateChain(ctx).RotateRequest(rotateRequest).Execute()

Rotate a chain

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
	rotateRequest := *openapiclient.NewRotateChainRequest() // RotateChainRequest | RotateRequest (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ChainsAPI.RotateChain(context.Background()).RotateRequest(rotateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.RotateChain``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRotateChainRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **rotateRequest** | [**RotateChainRequest**](RotateChainRequest.md) | RotateRequest | 

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


## SetChainRecord

> SetChainRecord(ctx, chainID).ChainRecord(chainRecord).Execute()

Sets the chain record.

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
	chainID := "chainID_example" // string | ChainID (Hex Address)
	chainRecord := *openapiclient.NewChainRecord([]string{"AccessNodes_example"}, false) // ChainRecord | Chain Record

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ChainsAPI.SetChainRecord(context.Background(), chainID).ChainRecord(chainRecord).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.SetChainRecord``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chainID** | **string** | ChainID (Hex Address) | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetChainRecordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **chainRecord** | [**ChainRecord**](ChainRecord.md) | Chain Record | 

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


## V1ChainEvmPost

> V1ChainEvmPost(ctx).Execute()

Ethereum JSON-RPC

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
	r, err := apiClient.ChainsAPI.V1ChainEvmPost(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.V1ChainEvmPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiV1ChainEvmPostRequest struct via the builder pattern


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


## V1ChainEvmWsGet

> V1ChainEvmWsGet(ctx).Execute()

Ethereum JSON-RPC (Websocket transport)

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
	r, err := apiClient.ChainsAPI.V1ChainEvmWsGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.V1ChainEvmWsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiV1ChainEvmWsGetRequest struct via the builder pattern


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


## WaitForRequest

> ReceiptResponse WaitForRequest(ctx, requestID).TimeoutSeconds(timeoutSeconds).WaitForL1Confirmation(waitForL1Confirmation).Execute()

Wait until the given request has been processed by the node

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
	requestID := "requestID_example" // string | RequestID (Hex)
	timeoutSeconds := int32(56) // int32 | The timeout in seconds, maximum 60s (optional)
	waitForL1Confirmation := true // bool | Wait for the block to be confirmed on L1 (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ChainsAPI.WaitForRequest(context.Background(), requestID).TimeoutSeconds(timeoutSeconds).WaitForL1Confirmation(waitForL1Confirmation).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ChainsAPI.WaitForRequest``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WaitForRequest`: ReceiptResponse
	fmt.Fprintf(os.Stdout, "Response from `ChainsAPI.WaitForRequest`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**requestID** | **string** | RequestID (Hex) | 

### Other Parameters

Other parameters are passed through a pointer to a apiWaitForRequestRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **timeoutSeconds** | **int32** | The timeout in seconds, maximum 60s | 
 **waitForL1Confirmation** | **bool** | Wait for the block to be confirmed on L1 | 

### Return type

[**ReceiptResponse**](ReceiptResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

