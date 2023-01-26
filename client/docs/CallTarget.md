# CallTarget

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Contract** | Pointer to **int32** |  | [optional] 
**EntryPoint** | Pointer to **int32** |  | [optional] 

## Methods

### NewCallTarget

`func NewCallTarget() *CallTarget`

NewCallTarget instantiates a new CallTarget object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCallTargetWithDefaults

`func NewCallTargetWithDefaults() *CallTarget`

NewCallTargetWithDefaults instantiates a new CallTarget object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContract

`func (o *CallTarget) GetContract() int32`

GetContract returns the Contract field if non-nil, zero value otherwise.

### GetContractOk

`func (o *CallTarget) GetContractOk() (*int32, bool)`

GetContractOk returns a tuple with the Contract field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContract

`func (o *CallTarget) SetContract(v int32)`

SetContract sets Contract field to given value.

### HasContract

`func (o *CallTarget) HasContract() bool`

HasContract returns a boolean if a field has been set.

### GetEntryPoint

`func (o *CallTarget) GetEntryPoint() int32`

GetEntryPoint returns the EntryPoint field if non-nil, zero value otherwise.

### GetEntryPointOk

`func (o *CallTarget) GetEntryPointOk() (*int32, bool)`

GetEntryPointOk returns a tuple with the EntryPoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntryPoint

`func (o *CallTarget) SetEntryPoint(v int32)`

SetEntryPoint sets EntryPoint field to given value.

### HasEntryPoint

`func (o *CallTarget) HasEntryPoint() bool`

HasEntryPoint returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


