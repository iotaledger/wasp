# GasFeePolicy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GasFeeTokenId** | Pointer to **string** | The gas fee token id. Empty if base token. | [optional] 
**GasPerToken** | Pointer to **int64** | The amount of gas per token. | [optional] 
**ValidatorFeeShare** | Pointer to **int32** | The validator fee share. | [optional] 

## Methods

### NewGasFeePolicy

`func NewGasFeePolicy() *GasFeePolicy`

NewGasFeePolicy instantiates a new GasFeePolicy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGasFeePolicyWithDefaults

`func NewGasFeePolicyWithDefaults() *GasFeePolicy`

NewGasFeePolicyWithDefaults instantiates a new GasFeePolicy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGasFeeTokenId

`func (o *GasFeePolicy) GetGasFeeTokenId() string`

GetGasFeeTokenId returns the GasFeeTokenId field if non-nil, zero value otherwise.

### GetGasFeeTokenIdOk

`func (o *GasFeePolicy) GetGasFeeTokenIdOk() (*string, bool)`

GetGasFeeTokenIdOk returns a tuple with the GasFeeTokenId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeeTokenId

`func (o *GasFeePolicy) SetGasFeeTokenId(v string)`

SetGasFeeTokenId sets GasFeeTokenId field to given value.

### HasGasFeeTokenId

`func (o *GasFeePolicy) HasGasFeeTokenId() bool`

HasGasFeeTokenId returns a boolean if a field has been set.

### GetGasPerToken

`func (o *GasFeePolicy) GetGasPerToken() int64`

GetGasPerToken returns the GasPerToken field if non-nil, zero value otherwise.

### GetGasPerTokenOk

`func (o *GasFeePolicy) GetGasPerTokenOk() (*int64, bool)`

GetGasPerTokenOk returns a tuple with the GasPerToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasPerToken

`func (o *GasFeePolicy) SetGasPerToken(v int64)`

SetGasPerToken sets GasPerToken field to given value.

### HasGasPerToken

`func (o *GasFeePolicy) HasGasPerToken() bool`

HasGasPerToken returns a boolean if a field has been set.

### GetValidatorFeeShare

`func (o *GasFeePolicy) GetValidatorFeeShare() int32`

GetValidatorFeeShare returns the ValidatorFeeShare field if non-nil, zero value otherwise.

### GetValidatorFeeShareOk

`func (o *GasFeePolicy) GetValidatorFeeShareOk() (*int32, bool)`

GetValidatorFeeShareOk returns a tuple with the ValidatorFeeShare field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidatorFeeShare

`func (o *GasFeePolicy) SetValidatorFeeShare(v int32)`

SetValidatorFeeShare sets ValidatorFeeShare field to given value.

### HasValidatorFeeShare

`func (o *GasFeePolicy) HasValidatorFeeShare() bool`

HasValidatorFeeShare returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


