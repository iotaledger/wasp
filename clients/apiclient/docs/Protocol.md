# Protocol

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Epoch** | [**BigInt**](BigInt.md) |  | 
**EpochDurationMs** | [**BigInt**](BigInt.md) |  | 
**EpochStartTimestampMs** | [**BigInt**](BigInt.md) |  | 
**ProtocolVersion** | [**BigInt**](BigInt.md) |  | 
**ReferenceGasPrice** | [**BigInt**](BigInt.md) |  | 
**SystemStateVersion** | [**BigInt**](BigInt.md) |  | 

## Methods

### NewProtocol

`func NewProtocol(epoch BigInt, epochDurationMs BigInt, epochStartTimestampMs BigInt, protocolVersion BigInt, referenceGasPrice BigInt, systemStateVersion BigInt, ) *Protocol`

NewProtocol instantiates a new Protocol object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtocolWithDefaults

`func NewProtocolWithDefaults() *Protocol`

NewProtocolWithDefaults instantiates a new Protocol object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEpoch

`func (o *Protocol) GetEpoch() BigInt`

GetEpoch returns the Epoch field if non-nil, zero value otherwise.

### GetEpochOk

`func (o *Protocol) GetEpochOk() (*BigInt, bool)`

GetEpochOk returns a tuple with the Epoch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEpoch

`func (o *Protocol) SetEpoch(v BigInt)`

SetEpoch sets Epoch field to given value.


### GetEpochDurationMs

`func (o *Protocol) GetEpochDurationMs() BigInt`

GetEpochDurationMs returns the EpochDurationMs field if non-nil, zero value otherwise.

### GetEpochDurationMsOk

`func (o *Protocol) GetEpochDurationMsOk() (*BigInt, bool)`

GetEpochDurationMsOk returns a tuple with the EpochDurationMs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEpochDurationMs

`func (o *Protocol) SetEpochDurationMs(v BigInt)`

SetEpochDurationMs sets EpochDurationMs field to given value.


### GetEpochStartTimestampMs

`func (o *Protocol) GetEpochStartTimestampMs() BigInt`

GetEpochStartTimestampMs returns the EpochStartTimestampMs field if non-nil, zero value otherwise.

### GetEpochStartTimestampMsOk

`func (o *Protocol) GetEpochStartTimestampMsOk() (*BigInt, bool)`

GetEpochStartTimestampMsOk returns a tuple with the EpochStartTimestampMs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEpochStartTimestampMs

`func (o *Protocol) SetEpochStartTimestampMs(v BigInt)`

SetEpochStartTimestampMs sets EpochStartTimestampMs field to given value.


### GetProtocolVersion

`func (o *Protocol) GetProtocolVersion() BigInt`

GetProtocolVersion returns the ProtocolVersion field if non-nil, zero value otherwise.

### GetProtocolVersionOk

`func (o *Protocol) GetProtocolVersionOk() (*BigInt, bool)`

GetProtocolVersionOk returns a tuple with the ProtocolVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocolVersion

`func (o *Protocol) SetProtocolVersion(v BigInt)`

SetProtocolVersion sets ProtocolVersion field to given value.


### GetReferenceGasPrice

`func (o *Protocol) GetReferenceGasPrice() BigInt`

GetReferenceGasPrice returns the ReferenceGasPrice field if non-nil, zero value otherwise.

### GetReferenceGasPriceOk

`func (o *Protocol) GetReferenceGasPriceOk() (*BigInt, bool)`

GetReferenceGasPriceOk returns a tuple with the ReferenceGasPrice field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReferenceGasPrice

`func (o *Protocol) SetReferenceGasPrice(v BigInt)`

SetReferenceGasPrice sets ReferenceGasPrice field to given value.


### GetSystemStateVersion

`func (o *Protocol) GetSystemStateVersion() BigInt`

GetSystemStateVersion returns the SystemStateVersion field if non-nil, zero value otherwise.

### GetSystemStateVersionOk

`func (o *Protocol) GetSystemStateVersionOk() (*BigInt, bool)`

GetSystemStateVersionOk returns a tuple with the SystemStateVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemStateVersion

`func (o *Protocol) SetSystemStateVersion(v BigInt)`

SetSystemStateVersion sets SystemStateVersion field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


