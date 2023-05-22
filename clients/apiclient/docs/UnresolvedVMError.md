# UnresolvedVMError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to [**VMErrorCode**](VMErrorCode.md) |  | [optional] 
**Hash** | Pointer to **int32** |  | [optional] 
**Params** | Pointer to **[]string** |  | [optional] 

## Methods

### NewUnresolvedVMError

`func NewUnresolvedVMError() *UnresolvedVMError`

NewUnresolvedVMError instantiates a new UnresolvedVMError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUnresolvedVMErrorWithDefaults

`func NewUnresolvedVMErrorWithDefaults() *UnresolvedVMError`

NewUnresolvedVMErrorWithDefaults instantiates a new UnresolvedVMError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *UnresolvedVMError) GetCode() VMErrorCode`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *UnresolvedVMError) GetCodeOk() (*VMErrorCode, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *UnresolvedVMError) SetCode(v VMErrorCode)`

SetCode sets Code field to given value.

### HasCode

`func (o *UnresolvedVMError) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetHash

`func (o *UnresolvedVMError) GetHash() int32`

GetHash returns the Hash field if non-nil, zero value otherwise.

### GetHashOk

`func (o *UnresolvedVMError) GetHashOk() (*int32, bool)`

GetHashOk returns a tuple with the Hash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHash

`func (o *UnresolvedVMError) SetHash(v int32)`

SetHash sets Hash field to given value.

### HasHash

`func (o *UnresolvedVMError) HasHash() bool`

HasHash returns a boolean if a field has been set.

### GetParams

`func (o *UnresolvedVMError) GetParams() []string`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *UnresolvedVMError) GetParamsOk() (*[]string, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *UnresolvedVMError) SetParams(v []string)`

SetParams sets Params field to given value.

### HasParams

`func (o *UnresolvedVMError) HasParams() bool`

HasParams returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


