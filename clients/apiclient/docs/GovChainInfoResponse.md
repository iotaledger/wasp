# GovChainInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainID** | **string** | ChainID (Bech32-encoded). | 
**ChainOwnerId** | **string** | The chain owner address (Bech32-encoded). | 
**Description** | **string** | The description of the chain. | 
**GasFeePolicy** | [**GasFeePolicy**](GasFeePolicy.md) |  | 
**MaxBlobSize** | **uint32** | The maximum contract blob size. | 
**MaxEventSize** | **uint32** | The maximum event size. | 
**MaxEventsPerReq** | **uint32** | The maximum amount of events per request. | 

## Methods

### NewGovChainInfoResponse

`func NewGovChainInfoResponse(chainID string, chainOwnerId string, description string, gasFeePolicy GasFeePolicy, maxBlobSize uint32, maxEventSize uint32, maxEventsPerReq uint32, ) *GovChainInfoResponse`

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


### GetMaxBlobSize

`func (o *GovChainInfoResponse) GetMaxBlobSize() uint32`

GetMaxBlobSize returns the MaxBlobSize field if non-nil, zero value otherwise.

### GetMaxBlobSizeOk

`func (o *GovChainInfoResponse) GetMaxBlobSizeOk() (*uint32, bool)`

GetMaxBlobSizeOk returns a tuple with the MaxBlobSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxBlobSize

`func (o *GovChainInfoResponse) SetMaxBlobSize(v uint32)`

SetMaxBlobSize sets MaxBlobSize field to given value.


### GetMaxEventSize

`func (o *GovChainInfoResponse) GetMaxEventSize() uint32`

GetMaxEventSize returns the MaxEventSize field if non-nil, zero value otherwise.

### GetMaxEventSizeOk

`func (o *GovChainInfoResponse) GetMaxEventSizeOk() (*uint32, bool)`

GetMaxEventSizeOk returns a tuple with the MaxEventSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventSize

`func (o *GovChainInfoResponse) SetMaxEventSize(v uint32)`

SetMaxEventSize sets MaxEventSize field to given value.


### GetMaxEventsPerReq

`func (o *GovChainInfoResponse) GetMaxEventsPerReq() uint32`

GetMaxEventsPerReq returns the MaxEventsPerReq field if non-nil, zero value otherwise.

### GetMaxEventsPerReqOk

`func (o *GovChainInfoResponse) GetMaxEventsPerReqOk() (*uint32, bool)`

GetMaxEventsPerReqOk returns a tuple with the MaxEventsPerReq field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxEventsPerReq

`func (o *GovChainInfoResponse) SetMaxEventsPerReq(v uint32)`

SetMaxEventsPerReq sets MaxEventsPerReq field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


