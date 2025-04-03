# ProtocolParameters

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Epoch** | **int64** | The protocol&#39;s current epoch | 
**EpochDurationMs** | **int64** | The current epoch&#39;s duration in ms | 
**EpochStartTimestampMs** | **int64** | The current epoch&#39;s start_timestamp in ms | 
**IotaTotalSupply** | **int64** | The iota&#39;s total_supply | 
**ProtocolVersion** | **int64** | The protocol&#39;s version | 
**ReferenceGasPrice** | **int64** | The current reference_gas_price | 
**SystemStateVersion** | **int64** | The protocol&#39;s system_state_version | 

## Methods

### NewProtocolParameters

`func NewProtocolParameters(epoch int64, epochDurationMs int64, epochStartTimestampMs int64, iotaTotalSupply int64, protocolVersion int64, referenceGasPrice int64, systemStateVersion int64, ) *ProtocolParameters`

NewProtocolParameters instantiates a new ProtocolParameters object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtocolParametersWithDefaults

`func NewProtocolParametersWithDefaults() *ProtocolParameters`

NewProtocolParametersWithDefaults instantiates a new ProtocolParameters object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEpoch

`func (o *ProtocolParameters) GetEpoch() int64`

GetEpoch returns the Epoch field if non-nil, zero value otherwise.

### GetEpochOk

`func (o *ProtocolParameters) GetEpochOk() (*int64, bool)`

GetEpochOk returns a tuple with the Epoch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEpoch

`func (o *ProtocolParameters) SetEpoch(v int64)`

SetEpoch sets Epoch field to given value.


### GetEpochDurationMs

`func (o *ProtocolParameters) GetEpochDurationMs() int64`

GetEpochDurationMs returns the EpochDurationMs field if non-nil, zero value otherwise.

### GetEpochDurationMsOk

`func (o *ProtocolParameters) GetEpochDurationMsOk() (*int64, bool)`

GetEpochDurationMsOk returns a tuple with the EpochDurationMs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEpochDurationMs

`func (o *ProtocolParameters) SetEpochDurationMs(v int64)`

SetEpochDurationMs sets EpochDurationMs field to given value.


### GetEpochStartTimestampMs

`func (o *ProtocolParameters) GetEpochStartTimestampMs() int64`

GetEpochStartTimestampMs returns the EpochStartTimestampMs field if non-nil, zero value otherwise.

### GetEpochStartTimestampMsOk

`func (o *ProtocolParameters) GetEpochStartTimestampMsOk() (*int64, bool)`

GetEpochStartTimestampMsOk returns a tuple with the EpochStartTimestampMs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEpochStartTimestampMs

`func (o *ProtocolParameters) SetEpochStartTimestampMs(v int64)`

SetEpochStartTimestampMs sets EpochStartTimestampMs field to given value.


### GetIotaTotalSupply

`func (o *ProtocolParameters) GetIotaTotalSupply() int64`

GetIotaTotalSupply returns the IotaTotalSupply field if non-nil, zero value otherwise.

### GetIotaTotalSupplyOk

`func (o *ProtocolParameters) GetIotaTotalSupplyOk() (*int64, bool)`

GetIotaTotalSupplyOk returns a tuple with the IotaTotalSupply field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIotaTotalSupply

`func (o *ProtocolParameters) SetIotaTotalSupply(v int64)`

SetIotaTotalSupply sets IotaTotalSupply field to given value.


### GetProtocolVersion

`func (o *ProtocolParameters) GetProtocolVersion() int64`

GetProtocolVersion returns the ProtocolVersion field if non-nil, zero value otherwise.

### GetProtocolVersionOk

`func (o *ProtocolParameters) GetProtocolVersionOk() (*int64, bool)`

GetProtocolVersionOk returns a tuple with the ProtocolVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocolVersion

`func (o *ProtocolParameters) SetProtocolVersion(v int64)`

SetProtocolVersion sets ProtocolVersion field to given value.


### GetReferenceGasPrice

`func (o *ProtocolParameters) GetReferenceGasPrice() int64`

GetReferenceGasPrice returns the ReferenceGasPrice field if non-nil, zero value otherwise.

### GetReferenceGasPriceOk

`func (o *ProtocolParameters) GetReferenceGasPriceOk() (*int64, bool)`

GetReferenceGasPriceOk returns a tuple with the ReferenceGasPrice field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReferenceGasPrice

`func (o *ProtocolParameters) SetReferenceGasPrice(v int64)`

SetReferenceGasPrice sets ReferenceGasPrice field to given value.


### GetSystemStateVersion

`func (o *ProtocolParameters) GetSystemStateVersion() int64`

GetSystemStateVersion returns the SystemStateVersion field if non-nil, zero value otherwise.

### GetSystemStateVersionOk

`func (o *ProtocolParameters) GetSystemStateVersionOk() (*int64, bool)`

GetSystemStateVersionOk returns a tuple with the SystemStateVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemStateVersion

`func (o *ProtocolParameters) SetSystemStateVersion(v int64)`

SetSystemStateVersion sets SystemStateVersion field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


