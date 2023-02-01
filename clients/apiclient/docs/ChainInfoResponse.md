# ChainInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainID** | **string** | ChainID (Bech32-encoded). | 
**ChainOwnerId** | **string** | The chain owner address (Bech32-encoded). | 
**Description** | **string** | The description of the chain. | 
**EvmChainId** | **uint32** | The EVM chain ID | 
**GasFeePolicy** | Pointer to [**GasFeePolicy**](GasFeePolicy.md) |  | [optional] 
**IsActive** | **bool** | Whether or not the chain is active. | 
**MaxBlobSize** | **uint32** | The maximum contract blob size. | 
**MaxEventSize** | **uint32** | The maximum event size. | 
**MaxEventsPerReq** | **uint32** | The maximum amount of events per request. | 

## Methods

### NewChainInfoResponse

`func NewChainInfoResponse(chainID string, chainOwnerId string, description string, evmChainId uint32, isActive bool, maxBlobSize uint32, maxEventSize uint32, maxEventsPerReq uint32, ) *ChainInfoResponse`

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


### GetEvmChainId

`func (o *ChainInfoResponse) GetEvmChainId() uint32`

GetEvmChainId returns the EvmChainId field if non-nil, zero value otherwise.

### GetEvmChainIdOk

`func (o *ChainInfoResponse) GetEvmChainIdOk() (*uint32, bool)`

GetEvmChainIdOk returns a tuple with the EvmChainId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmChainId

`func (o *ChainInfoResponse) SetEvmChainId(v uint32)`

SetEvmChainId sets EvmChainId field to given value.


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


### GetMaxBlobSize

`func (o *ChainInfoResponse) GetMaxBlobSize() uint32`

GetMaxBlobSize returns the MaxBlobSize field if non-nil, zero value otherwise.

### GetMaxBlobSizeOk

`func (o *ChainInfoResponse) GetMaxBlobSizeOk() (*uint32, bool)`

GetMaxBlobSizeOk returns a tuple with the MaxBlobSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxBlobSize

`func (o *ChainInfoResponse) SetMaxBlobSize(v uint32)`

SetMaxBlobSize sets MaxBlobSize field to given value.


### GetMaxEventSize

`func (o *ChainInfoResponse) GetMaxEventSize() uint32`

GetMaxEventSize returns the MaxEventSize field if non-nil, zero value otherwise.

### GetMaxEventSizeOk

`func (o *ChainInfoResponse) GetMaxEventSizeOk() (*uint32, bool)`

GetMaxEventSizeOk returns a tuple with the MaxEventSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventSize

`func (o *ChainInfoResponse) SetMaxEventSize(v uint32)`

SetMaxEventSize sets MaxEventSize field to given value.


### GetMaxEventsPerReq

`func (o *ChainInfoResponse) GetMaxEventsPerReq() uint32`

GetMaxEventsPerReq returns the MaxEventsPerReq field if non-nil, zero value otherwise.

### GetMaxEventsPerReqOk

`func (o *ChainInfoResponse) GetMaxEventsPerReqOk() (*uint32, bool)`

GetMaxEventsPerReqOk returns a tuple with the MaxEventsPerReq field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventsPerReq

`func (o *ChainInfoResponse) SetMaxEventsPerReq(v uint32)`

SetMaxEventsPerReq sets MaxEventsPerReq field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


