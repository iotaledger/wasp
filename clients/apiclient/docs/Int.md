# Int

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Abs** | Pointer to **[]int32** |  | [optional] 
**Neg** | Pointer to **bool** |  | [optional] 

## Methods

### NewInt

`func NewInt() *Int`

NewInt instantiates a new Int object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIntWithDefaults

`func NewIntWithDefaults() *Int`

NewIntWithDefaults instantiates a new Int object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAbs

`func (o *Int) GetAbs() []int32`

GetAbs returns the Abs field if non-nil, zero value otherwise.

### GetAbsOk

`func (o *Int) GetAbsOk() (*[]int32, bool)`

GetAbsOk returns a tuple with the Abs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAbs

`func (o *Int) SetAbs(v []int32)`

SetAbs sets Abs field to given value.

### HasAbs

`func (o *Int) HasAbs() bool`

HasAbs returns a boolean if a field has been set.

### GetNeg

`func (o *Int) GetNeg() bool`

GetNeg returns the Neg field if non-nil, zero value otherwise.

### GetNegOk

`func (o *Int) GetNegOk() (*bool, bool)`

GetNegOk returns a tuple with the Neg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNeg

`func (o *Int) SetNeg(v bool)`

SetNeg sets Neg field to given value.

### HasNeg

`func (o *Int) HasNeg() bool`

HasNeg returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


