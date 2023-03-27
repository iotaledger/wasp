# Limits

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MaxGasExternalViewCall** | **int64** | The maximum gas per external view call | 
**MaxGasPerBlock** | **int64** | The maximum gas per block | 
**MaxGasPerRequest** | **int64** | The maximum gas per request | 
**MinGasPerRequest** | **int64** | The minimum gas per request | 

## Methods

### NewLimits

`func NewLimits(maxGasExternalViewCall int64, maxGasPerBlock int64, maxGasPerRequest int64, minGasPerRequest int64, ) *Limits`

NewLimits instantiates a new Limits object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLimitsWithDefaults

`func NewLimitsWithDefaults() *Limits`

NewLimitsWithDefaults instantiates a new Limits object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMaxGasExternalViewCall

`func (o *Limits) GetMaxGasExternalViewCall() int64`

GetMaxGasExternalViewCall returns the MaxGasExternalViewCall field if non-nil, zero value otherwise.

### GetMaxGasExternalViewCallOk

`func (o *Limits) GetMaxGasExternalViewCallOk() (*int64, bool)`

GetMaxGasExternalViewCallOk returns a tuple with the MaxGasExternalViewCall field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxGasExternalViewCall

`func (o *Limits) SetMaxGasExternalViewCall(v int64)`

SetMaxGasExternalViewCall sets MaxGasExternalViewCall field to given value.


### GetMaxGasPerBlock

`func (o *Limits) GetMaxGasPerBlock() int64`

GetMaxGasPerBlock returns the MaxGasPerBlock field if non-nil, zero value otherwise.

### GetMaxGasPerBlockOk

`func (o *Limits) GetMaxGasPerBlockOk() (*int64, bool)`

GetMaxGasPerBlockOk returns a tuple with the MaxGasPerBlock field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxGasPerBlock

`func (o *Limits) SetMaxGasPerBlock(v int64)`

SetMaxGasPerBlock sets MaxGasPerBlock field to given value.


### GetMaxGasPerRequest

`func (o *Limits) GetMaxGasPerRequest() int64`

GetMaxGasPerRequest returns the MaxGasPerRequest field if non-nil, zero value otherwise.

### GetMaxGasPerRequestOk

`func (o *Limits) GetMaxGasPerRequestOk() (*int64, bool)`

GetMaxGasPerRequestOk returns a tuple with the MaxGasPerRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxGasPerRequest

`func (o *Limits) SetMaxGasPerRequest(v int64)`

SetMaxGasPerRequest sets MaxGasPerRequest field to given value.


### GetMinGasPerRequest

`func (o *Limits) GetMinGasPerRequest() int64`

GetMinGasPerRequest returns the MinGasPerRequest field if non-nil, zero value otherwise.

### GetMinGasPerRequestOk

`func (o *Limits) GetMinGasPerRequestOk() (*int64, bool)`

GetMinGasPerRequestOk returns a tuple with the MinGasPerRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinGasPerRequest

`func (o *Limits) SetMinGasPerRequest(v int64)`

SetMinGasPerRequest sets MinGasPerRequest field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


