# ProtocolParameters

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bech32Hrp** | **string** | The human readable network prefix | 
**BelowMaxDepth** | **uint32** | The networks max depth | 
**MinPowScore** | **uint32** | The minimal PoW score | 
**NetworkName** | **string** | The network name | 
**RentStructure** | [**RentStructure**](RentStructure.md) |  | 
**TokenSupply** | **string** | The token supply | 
**Version** | **uint32** | The protocol version | 

## Methods

### NewProtocolParameters

`func NewProtocolParameters(bech32Hrp string, belowMaxDepth uint32, minPowScore uint32, networkName string, rentStructure RentStructure, tokenSupply string, version uint32, ) *ProtocolParameters`

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


### GetBelowMaxDepth

`func (o *ProtocolParameters) GetBelowMaxDepth() uint32`

GetBelowMaxDepth returns the BelowMaxDepth field if non-nil, zero value otherwise.

### GetBelowMaxDepthOk

`func (o *ProtocolParameters) GetBelowMaxDepthOk() (*uint32, bool)`

GetBelowMaxDepthOk returns a tuple with the BelowMaxDepth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBelowMaxDepth

`func (o *ProtocolParameters) SetBelowMaxDepth(v uint32)`

SetBelowMaxDepth sets BelowMaxDepth field to given value.


### GetMinPowScore

`func (o *ProtocolParameters) GetMinPowScore() uint32`

GetMinPowScore returns the MinPowScore field if non-nil, zero value otherwise.

### GetMinPowScoreOk

`func (o *ProtocolParameters) GetMinPowScoreOk() (*uint32, bool)`

GetMinPowScoreOk returns a tuple with the MinPowScore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinPowScore

`func (o *ProtocolParameters) SetMinPowScore(v uint32)`

SetMinPowScore sets MinPowScore field to given value.


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


### GetVersion

`func (o *ProtocolParameters) GetVersion() uint32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ProtocolParameters) GetVersionOk() (*uint32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ProtocolParameters) SetVersion(v uint32)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


