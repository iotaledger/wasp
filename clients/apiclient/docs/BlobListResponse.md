# BlobListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Blobs** | Pointer to [**[]Blob**](Blob.md) |  | [optional] 

## Methods

### NewBlobListResponse

`func NewBlobListResponse() *BlobListResponse`

NewBlobListResponse instantiates a new BlobListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBlobListResponseWithDefaults

`func NewBlobListResponseWithDefaults() *BlobListResponse`

NewBlobListResponseWithDefaults instantiates a new BlobListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBlobs

`func (o *BlobListResponse) GetBlobs() []Blob`

GetBlobs returns the Blobs field if non-nil, zero value otherwise.

### GetBlobsOk

`func (o *BlobListResponse) GetBlobsOk() (*[]Blob, bool)`

GetBlobsOk returns a tuple with the Blobs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlobs

`func (o *BlobListResponse) SetBlobs(v []Blob)`

SetBlobs sets Blobs field to given value.

### HasBlobs

`func (o *BlobListResponse) HasBlobs() bool`

HasBlobs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


