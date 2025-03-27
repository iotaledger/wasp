# \RequestsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**OffLedger**](RequestsAPI.md#OffLedger) | **Post** /v1/requests/offledger | Post an off-ledger request



## OffLedger

> OffLedger(ctx).OffLedgerRequest(offLedgerRequest).Execute()

Post an off-ledger request

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
	offLedgerRequest := *openapiclient.NewOffLedgerRequest("Hex string") // OffLedgerRequest | Offledger request as JSON. Request encoded in Hex

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RequestsAPI.OffLedger(context.Background()).OffLedgerRequest(offLedgerRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RequestsAPI.OffLedger``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOffLedgerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offLedgerRequest** | [**OffLedgerRequest**](OffLedgerRequest.md) | Offledger request as JSON. Request encoded in Hex | 

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

