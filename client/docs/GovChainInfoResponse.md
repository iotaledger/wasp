# GovChainInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainID** | Pointer to **string** | ChainID (Bech32-encoded). | [optional] 
**ChainOwnerId** | Pointer to **string** | The chain owner address (Bech32-encoded). | [optional] 
**Description** | Pointer to **string** | The description of the chain. | [optional] 
**GasFeePolicy** | Pointer to [**GasFeePolicy**](GasFeePolicy.md) |  | [optional] 
**MaxBlobSize** | Pointer to **int32** | The maximum contract blob size. | [optional] 
**MaxEventSize** | Pointer to **int32** | The maximum event size. | [optional] 
**MaxEventsPerReq** | Pointer to **int32** | The maximum amount of events per request. | [optional] 

## Methods

### NewGovChainInfoResponse

`func NewGovChainInfoResponse() *GovChainInfoResponse`

NewGovChainInfoResponse instantiates a new GovChainInfoResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGovChainInfoResponseWithDefaults

`func NewGovChainInfoResponseWithDefaults() *GovChainInfoResponse`

NewGovChainInfoResponseWithDefaults instantiates a new GovChainInfoResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChainID

`func (o *GovChainInfoResponse) GetChainID() string`

GetChainID returns the ChainID field if non-nil, zero value otherwise.

### GetChainIDOk

`func (o *GovChainInfoResponse) GetChainIDOk() (*string, bool)`

GetChainIDOk returns a tuple with the ChainID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainID

`func (o *GovChainInfoResponse) SetChainID(v string)`

SetChainID sets ChainID field to given value.

### HasChainID

`func (o *GovChainInfoResponse) HasChainID() bool`

HasChainID returns a boolean if a field has been set.

### GetChainOwnerId

`func (o *GovChainInfoResponse) GetChainOwnerId() string`

GetChainOwnerId returns the ChainOwnerId field if non-nil, zero value otherwise.

### GetChainOwnerIdOk

`func (o *GovChainInfoResponse) GetChainOwnerIdOk() (*string, bool)`

GetChainOwnerIdOk returns a tuple with the ChainOwnerId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainOwnerId

`func (o *GovChainInfoResponse) SetChainOwnerId(v string)`

SetChainOwnerId sets ChainOwnerId field to given value.

### HasChainOwnerId

`func (o *GovChainInfoResponse) HasChainOwnerId() bool`

HasChainOwnerId returns a boolean if a field has been set.

### GetDescription

`func (o *GovChainInfoResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GovChainInfoResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GovChainInfoResponse) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GovChainInfoResponse) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetGasFeePolicy

`func (o *GovChainInfoResponse) GetGasFeePolicy() GasFeePolicy`

GetGasFeePolicy returns the GasFeePolicy field if non-nil, zero value otherwise.

### GetGasFeePolicyOk

`func (o *GovChainInfoResponse) GetGasFeePolicyOk() (*GasFeePolicy, bool)`

GetGasFeePolicyOk returns a tuple with the GasFeePolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeePolicy

`func (o *GovChainInfoResponse) SetGasFeePolicy(v GasFeePolicy)`

SetGasFeePolicy sets GasFeePolicy field to given value.

### HasGasFeePolicy

`func (o *GovChainInfoResponse) HasGasFeePolicy() bool`

HasGasFeePolicy returns a boolean if a field has been set.

### GetMaxBlobSize

`func (o *GovChainInfoResponse) GetMaxBlobSize() int32`

GetMaxBlobSize returns the MaxBlobSize field if non-nil, zero value otherwise.

### GetMaxBlobSizeOk

`func (o *GovChainInfoResponse) GetMaxBlobSizeOk() (*int32, bool)`

GetMaxBlobSizeOk returns a tuple with the MaxBlobSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxBlobSize

`func (o *GovChainInfoResponse) SetMaxBlobSize(v int32)`

SetMaxBlobSize sets MaxBlobSize field to given value.

### HasMaxBlobSize

`func (o *GovChainInfoResponse) HasMaxBlobSize() bool`

HasMaxBlobSize returns a boolean if a field has been set.

### GetMaxEventSize

`func (o *GovChainInfoResponse) GetMaxEventSize() int32`

GetMaxEventSize returns the MaxEventSize field if non-nil, zero value otherwise.

### GetMaxEventSizeOk

`func (o *GovChainInfoResponse) GetMaxEventSizeOk() (*int32, bool)`

GetMaxEventSizeOk returns a tuple with the MaxEventSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventSize

`func (o *GovChainInfoResponse) SetMaxEventSize(v int32)`

SetMaxEventSize sets MaxEventSize field to given value.

### HasMaxEventSize

`func (o *GovChainInfoResponse) HasMaxEventSize() bool`

HasMaxEventSize returns a boolean if a field has been set.

### GetMaxEventsPerReq

`func (o *GovChainInfoResponse) GetMaxEventsPerReq() int32`

GetMaxEventsPerReq returns the MaxEventsPerReq field if non-nil, zero value otherwise.

### GetMaxEventsPerReqOk

`func (o *GovChainInfoResponse) GetMaxEventsPerReqOk() (*int32, bool)`

GetMaxEventsPerReqOk returns a tuple with the MaxEventsPerReq field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventsPerReq

`func (o *GovChainInfoResponse) SetMaxEventsPerReq(v int32)`

SetMaxEventsPerReq sets MaxEventsPerReq field to given value.

### HasMaxEventsPerReq

`func (o *GovChainInfoResponse) HasMaxEventsPerReq() bool`

HasMaxEventsPerReq returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


