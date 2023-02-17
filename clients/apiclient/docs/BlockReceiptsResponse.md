# BlockReceiptsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Receipts** | [**[]RequestReceiptResponse**](RequestReceiptResponse.md) |  | 

## Methods

### NewBlockReceiptsResponse

`func NewBlockReceiptsResponse(receipts []RequestReceiptResponse, ) *BlockReceiptsResponse`

NewBlockReceiptsResponse instantiates a new BlockReceiptsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBlockReceiptsResponseWithDefaults

`func NewBlockReceiptsResponseWithDefaults() *BlockReceiptsResponse`

NewBlockReceiptsResponseWithDefaults instantiates a new BlockReceiptsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReceipts

`func (o *BlockReceiptsResponse) GetReceipts() []RequestReceiptResponse`

GetReceipts returns the Receipts field if non-nil, zero value otherwise.

### GetReceiptsOk

`func (o *BlockReceiptsResponse) GetReceiptsOk() (*[]RequestReceiptResponse, bool)`

GetReceiptsOk returns a tuple with the Receipts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReceipts

`func (o *BlockReceiptsResponse) SetReceipts(v []RequestReceiptResponse)`

SetReceipts sets Receipts field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


