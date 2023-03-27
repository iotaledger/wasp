# FeePolicy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EvmGasRatio** | [**Ratio32**](Ratio32.md) |  | 
**GasPerToken** | [**Ratio32**](Ratio32.md) |  | 
**ValidatorFeeShare** | **int32** | The validator fee share. | 

## Methods

### NewFeePolicy

`func NewFeePolicy(evmGasRatio Ratio32, gasPerToken Ratio32, validatorFeeShare int32, ) *FeePolicy`

NewFeePolicy instantiates a new FeePolicy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFeePolicyWithDefaults

`func NewFeePolicyWithDefaults() *FeePolicy`

NewFeePolicyWithDefaults instantiates a new FeePolicy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEvmGasRatio

`func (o *FeePolicy) GetEvmGasRatio() Ratio32`

GetEvmGasRatio returns the EvmGasRatio field if non-nil, zero value otherwise.

### GetEvmGasRatioOk

`func (o *FeePolicy) GetEvmGasRatioOk() (*Ratio32, bool)`

GetEvmGasRatioOk returns a tuple with the EvmGasRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmGasRatio

`func (o *FeePolicy) SetEvmGasRatio(v Ratio32)`

SetEvmGasRatio sets EvmGasRatio field to given value.


### GetGasPerToken

`func (o *FeePolicy) GetGasPerToken() Ratio32`

GetGasPerToken returns the GasPerToken field if non-nil, zero value otherwise.

### GetGasPerTokenOk

`func (o *FeePolicy) GetGasPerTokenOk() (*Ratio32, bool)`

GetGasPerTokenOk returns a tuple with the GasPerToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasPerToken

`func (o *FeePolicy) SetGasPerToken(v Ratio32)`

SetGasPerToken sets GasPerToken field to given value.


### GetValidatorFeeShare

`func (o *FeePolicy) GetValidatorFeeShare() int32`

GetValidatorFeeShare returns the ValidatorFeeShare field if non-nil, zero value otherwise.

### GetValidatorFeeShareOk

`func (o *FeePolicy) GetValidatorFeeShareOk() (*int32, bool)`

GetValidatorFeeShareOk returns a tuple with the ValidatorFeeShare field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidatorFeeShare

`func (o *FeePolicy) SetValidatorFeeShare(v int32)`

SetValidatorFeeShare sets ValidatorFeeShare field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


