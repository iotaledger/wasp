# ProtocolParameters

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bech32Hrp** | Pointer to **string** | The human readable network prefix | [optional] 
**BelowMaxDepth** | Pointer to **int32** | The networks max depth | [optional] 
**MinPowScore** | Pointer to **int32** | The minimal PoW score | [optional] 
**NetworkName** | Pointer to **string** | The network name | [optional] 
**RentStructure** | Pointer to [**RentStructure**](RentStructure.md) |  | [optional] 
**TokenSupply** | Pointer to **string** | The token supply | [optional] 
**Version** | Pointer to **int32** | The protocol version | [optional] 

## Methods

### NewProtocolParameters

`func NewProtocolParameters() *ProtocolParameters`

NewProtocolParameters instantiates a new ProtocolParameters object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtocolParametersWithDefaults

`func NewProtocolParametersWithDefaults() *ProtocolParameters`

NewProtocolParametersWithDefaults instantiates a new ProtocolParameters object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBech32Hrp

`func (o *ProtocolParameters) GetBech32Hrp() string`

GetBech32Hrp returns the Bech32Hrp field if non-nil, zero value otherwise.

### GetBech32HrpOk

`func (o *ProtocolParameters) GetBech32HrpOk() (*string, bool)`

GetBech32HrpOk returns a tuple with the Bech32Hrp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBech32Hrp

`func (o *ProtocolParameters) SetBech32Hrp(v string)`

SetBech32Hrp sets Bech32Hrp field to given value.

### HasBech32Hrp

`func (o *ProtocolParameters) HasBech32Hrp() bool`

HasBech32Hrp returns a boolean if a field has been set.

### GetBelowMaxDepth

`func (o *ProtocolParameters) GetBelowMaxDepth() int32`

GetBelowMaxDepth returns the BelowMaxDepth field if non-nil, zero value otherwise.

### GetBelowMaxDepthOk

`func (o *ProtocolParameters) GetBelowMaxDepthOk() (*int32, bool)`

GetBelowMaxDepthOk returns a tuple with the BelowMaxDepth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBelowMaxDepth

`func (o *ProtocolParameters) SetBelowMaxDepth(v int32)`

SetBelowMaxDepth sets BelowMaxDepth field to given value.

### HasBelowMaxDepth

`func (o *ProtocolParameters) HasBelowMaxDepth() bool`

HasBelowMaxDepth returns a boolean if a field has been set.

### GetMinPowScore

`func (o *ProtocolParameters) GetMinPowScore() int32`

GetMinPowScore returns the MinPowScore field if non-nil, zero value otherwise.

### GetMinPowScoreOk

`func (o *ProtocolParameters) GetMinPowScoreOk() (*int32, bool)`

GetMinPowScoreOk returns a tuple with the MinPowScore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinPowScore

`func (o *ProtocolParameters) SetMinPowScore(v int32)`

SetMinPowScore sets MinPowScore field to given value.

### HasMinPowScore

`func (o *ProtocolParameters) HasMinPowScore() bool`

HasMinPowScore returns a boolean if a field has been set.

### GetNetworkName

`func (o *ProtocolParameters) GetNetworkName() string`

GetNetworkName returns the NetworkName field if non-nil, zero value otherwise.

### GetNetworkNameOk

`func (o *ProtocolParameters) GetNetworkNameOk() (*string, bool)`

GetNetworkNameOk returns a tuple with the NetworkName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkName

`func (o *ProtocolParameters) SetNetworkName(v string)`

SetNetworkName sets NetworkName field to given value.

### HasNetworkName

`func (o *ProtocolParameters) HasNetworkName() bool`

HasNetworkName returns a boolean if a field has been set.

### GetRentStructure

`func (o *ProtocolParameters) GetRentStructure() RentStructure`

GetRentStructure returns the RentStructure field if non-nil, zero value otherwise.

### GetRentStructureOk

`func (o *ProtocolParameters) GetRentStructureOk() (*RentStructure, bool)`

GetRentStructureOk returns a tuple with the RentStructure field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRentStructure

`func (o *ProtocolParameters) SetRentStructure(v RentStructure)`

SetRentStructure sets RentStructure field to given value.

### HasRentStructure

`func (o *ProtocolParameters) HasRentStructure() bool`

HasRentStructure returns a boolean if a field has been set.

### GetTokenSupply

`func (o *ProtocolParameters) GetTokenSupply() string`

GetTokenSupply returns the TokenSupply field if non-nil, zero value otherwise.

### GetTokenSupplyOk

`func (o *ProtocolParameters) GetTokenSupplyOk() (*string, bool)`

GetTokenSupplyOk returns a tuple with the TokenSupply field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenSupply

`func (o *ProtocolParameters) SetTokenSupply(v string)`

SetTokenSupply sets TokenSupply field to given value.

### HasTokenSupply

`func (o *ProtocolParameters) HasTokenSupply() bool`

HasTokenSupply returns a boolean if a field has been set.

### GetVersion

`func (o *ProtocolParameters) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ProtocolParameters) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ProtocolParameters) SetVersion(v int32)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *ProtocolParameters) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


