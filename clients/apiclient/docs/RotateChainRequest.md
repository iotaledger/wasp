# RotateChainRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RotateToAddress** | Pointer to **string** | The address of the new committee or empty to cancel attempt to rotate | [optional] 

## Methods

### NewRotateChainRequest

`func NewRotateChainRequest() *RotateChainRequest`

NewRotateChainRequest instantiates a new RotateChainRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRotateChainRequestWithDefaults

`func NewRotateChainRequestWithDefaults() *RotateChainRequest`

NewRotateChainRequestWithDefaults instantiates a new RotateChainRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRotateToAddress

`func (o *RotateChainRequest) GetRotateToAddress() string`

GetRotateToAddress returns the RotateToAddress field if non-nil, zero value otherwise.

### GetRotateToAddressOk

`func (o *RotateChainRequest) GetRotateToAddressOk() (*string, bool)`

GetRotateToAddressOk returns a tuple with the RotateToAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRotateToAddress

`func (o *RotateChainRequest) SetRotateToAddress(v string)`

SetRotateToAddress sets RotateToAddress field to given value.

### HasRotateToAddress

`func (o *RotateChainRequest) HasRotateToAddress() bool`

HasRotateToAddress returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


