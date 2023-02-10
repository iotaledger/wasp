# BlockReceiptError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ErrorMessage** | **string** |  | 
**Hash** | **string** |  | 

## Methods

### NewBlockReceiptError

`func NewBlockReceiptError(errorMessage string, hash string, ) *BlockReceiptError`

NewBlockReceiptError instantiates a new BlockReceiptError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBlockReceiptErrorWithDefaults

`func NewBlockReceiptErrorWithDefaults() *BlockReceiptError`

NewBlockReceiptErrorWithDefaults instantiates a new BlockReceiptError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetErrorMessage

`func (o *BlockReceiptError) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *BlockReceiptError) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *BlockReceiptError) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.


### GetHash

`func (o *BlockReceiptError) GetHash() string`

GetHash returns the Hash field if non-nil, zero value otherwise.

### GetHashOk

`func (o *BlockReceiptError) GetHashOk() (*string, bool)`

GetHashOk returns a tuple with the Hash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHash

`func (o *BlockReceiptError) SetHash(v string)`

SetHash sets Hash field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


