# Blob

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Hash** | **string** |  | 
**Size** | **uint32** |  | 

## Methods

### NewBlob

`func NewBlob(hash string, size uint32, ) *Blob`

NewBlob instantiates a new Blob object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBlobWithDefaults

`func NewBlobWithDefaults() *Blob`

NewBlobWithDefaults instantiates a new Blob object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHash

`func (o *Blob) GetHash() string`

GetHash returns the Hash field if non-nil, zero value otherwise.

### GetHashOk

`func (o *Blob) GetHashOk() (*string, bool)`

GetHashOk returns a tuple with the Hash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHash

`func (o *Blob) SetHash(v string)`

SetHash sets Hash field to given value.


### GetSize

`func (o *Blob) GetSize() uint32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *Blob) GetSizeOk() (*uint32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *Blob) SetSize(v uint32)`

SetSize sets Size field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


