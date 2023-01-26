# ChainInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainID** | Pointer to **string** | ChainID (Bech32-encoded). | [optional] 
**ChainOwnerId** | Pointer to **string** | The chain owner address (Bech32-encoded). | [optional] 
**Description** | Pointer to **string** | The description of the chain. | [optional] 
**EvmChainId** | Pointer to **int32** | The EVM chain ID | [optional] 
**GasFeePolicy** | Pointer to [**GasFeePolicy**](GasFeePolicy.md) |  | [optional] 
**IsActive** | Pointer to **bool** | Whether or not the chain is active. | [optional] 
**MaxBlobSize** | Pointer to **int32** | The maximum contract blob size. | [optional] 
**MaxEventSize** | Pointer to **int32** | The maximum event size. | [optional] 
**MaxEventsPerReq** | Pointer to **int32** | The maximum amount of events per request. | [optional] 

## Methods

### NewChainInfoResponse

`func NewChainInfoResponse() *ChainInfoResponse`

NewChainInfoResponse instantiates a new ChainInfoResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewChainInfoResponseWithDefaults

`func NewChainInfoResponseWithDefaults() *ChainInfoResponse`

NewChainInfoResponseWithDefaults instantiates a new ChainInfoResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChainID

`func (o *ChainInfoResponse) GetChainID() string`

GetChainID returns the ChainID field if non-nil, zero value otherwise.

### GetChainIDOk

`func (o *ChainInfoResponse) GetChainIDOk() (*string, bool)`

GetChainIDOk returns a tuple with the ChainID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainID

`func (o *ChainInfoResponse) SetChainID(v string)`

SetChainID sets ChainID field to given value.

### HasChainID

`func (o *ChainInfoResponse) HasChainID() bool`

HasChainID returns a boolean if a field has been set.

### GetChainOwnerId

`func (o *ChainInfoResponse) GetChainOwnerId() string`

GetChainOwnerId returns the ChainOwnerId field if non-nil, zero value otherwise.

### GetChainOwnerIdOk

`func (o *ChainInfoResponse) GetChainOwnerIdOk() (*string, bool)`

GetChainOwnerIdOk returns a tuple with the ChainOwnerId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainOwnerId

`func (o *ChainInfoResponse) SetChainOwnerId(v string)`

SetChainOwnerId sets ChainOwnerId field to given value.

### HasChainOwnerId

`func (o *ChainInfoResponse) HasChainOwnerId() bool`

HasChainOwnerId returns a boolean if a field has been set.

### GetDescription

`func (o *ChainInfoResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ChainInfoResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ChainInfoResponse) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ChainInfoResponse) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetEvmChainId

`func (o *ChainInfoResponse) GetEvmChainId() int32`

GetEvmChainId returns the EvmChainId field if non-nil, zero value otherwise.

### GetEvmChainIdOk

`func (o *ChainInfoResponse) GetEvmChainIdOk() (*int32, bool)`

GetEvmChainIdOk returns a tuple with the EvmChainId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmChainId

`func (o *ChainInfoResponse) SetEvmChainId(v int32)`

SetEvmChainId sets EvmChainId field to given value.

### HasEvmChainId

`func (o *ChainInfoResponse) HasEvmChainId() bool`

HasEvmChainId returns a boolean if a field has been set.

### GetGasFeePolicy

`func (o *ChainInfoResponse) GetGasFeePolicy() GasFeePolicy`

GetGasFeePolicy returns the GasFeePolicy field if non-nil, zero value otherwise.

### GetGasFeePolicyOk

`func (o *ChainInfoResponse) GetGasFeePolicyOk() (*GasFeePolicy, bool)`

GetGasFeePolicyOk returns a tuple with the GasFeePolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeePolicy

`func (o *ChainInfoResponse) SetGasFeePolicy(v GasFeePolicy)`

SetGasFeePolicy sets GasFeePolicy field to given value.

### HasGasFeePolicy

`func (o *ChainInfoResponse) HasGasFeePolicy() bool`

HasGasFeePolicy returns a boolean if a field has been set.

### GetIsActive

`func (o *ChainInfoResponse) GetIsActive() bool`

GetIsActive returns the IsActive field if non-nil, zero value otherwise.

### GetIsActiveOk

`func (o *ChainInfoResponse) GetIsActiveOk() (*bool, bool)`

GetIsActiveOk returns a tuple with the IsActive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsActive

`func (o *ChainInfoResponse) SetIsActive(v bool)`

SetIsActive sets IsActive field to given value.

### HasIsActive

`func (o *ChainInfoResponse) HasIsActive() bool`

HasIsActive returns a boolean if a field has been set.

### GetMaxBlobSize

`func (o *ChainInfoResponse) GetMaxBlobSize() int32`

GetMaxBlobSize returns the MaxBlobSize field if non-nil, zero value otherwise.

### GetMaxBlobSizeOk

`func (o *ChainInfoResponse) GetMaxBlobSizeOk() (*int32, bool)`

GetMaxBlobSizeOk returns a tuple with the MaxBlobSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxBlobSize

`func (o *ChainInfoResponse) SetMaxBlobSize(v int32)`

SetMaxBlobSize sets MaxBlobSize field to given value.

### HasMaxBlobSize

`func (o *ChainInfoResponse) HasMaxBlobSize() bool`

HasMaxBlobSize returns a boolean if a field has been set.

### GetMaxEventSize

`func (o *ChainInfoResponse) GetMaxEventSize() int32`

GetMaxEventSize returns the MaxEventSize field if non-nil, zero value otherwise.

### GetMaxEventSizeOk

`func (o *ChainInfoResponse) GetMaxEventSizeOk() (*int32, bool)`

GetMaxEventSizeOk returns a tuple with the MaxEventSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventSize

`func (o *ChainInfoResponse) SetMaxEventSize(v int32)`

SetMaxEventSize sets MaxEventSize field to given value.

### HasMaxEventSize

`func (o *ChainInfoResponse) HasMaxEventSize() bool`

HasMaxEventSize returns a boolean if a field has been set.

### GetMaxEventsPerReq

`func (o *ChainInfoResponse) GetMaxEventsPerReq() int32`

GetMaxEventsPerReq returns the MaxEventsPerReq field if non-nil, zero value otherwise.

### GetMaxEventsPerReqOk

`func (o *ChainInfoResponse) GetMaxEventsPerReqOk() (*int32, bool)`

GetMaxEventsPerReqOk returns a tuple with the MaxEventsPerReq field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventsPerReq

`func (o *ChainInfoResponse) SetMaxEventsPerReq(v int32)`

SetMaxEventsPerReq sets MaxEventsPerReq field to given value.

### HasMaxEventsPerReq

`func (o *ChainInfoResponse) HasMaxEventsPerReq() bool`

HasMaxEventsPerReq returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


