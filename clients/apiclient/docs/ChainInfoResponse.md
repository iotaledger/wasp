# ChainInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainID** | **string** | ChainID (Bech32-encoded). | 
**ChainOwnerId** | **string** | The chain owner address (Bech32-encoded). | 
**CustomMetadata** | Pointer to **string** | (base64) Optional extra metadata that is appended to the L1 AliasOutput | [optional] 
**EvmChainId** | **uint32** | The EVM chain ID | 
**GasFeePolicy** | Pointer to [**GasFeePolicy**](GasFeePolicy.md) |  | [optional] 
**IsActive** | **bool** | Whether or not the chain is active. | 

## Methods

### NewChainInfoResponse

`func NewChainInfoResponse(chainID string, chainOwnerId string, evmChainId uint32, isActive bool, ) *ChainInfoResponse`

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


### GetCustomMetadata

`func (o *ChainInfoResponse) GetCustomMetadata() string`

GetCustomMetadata returns the CustomMetadata field if non-nil, zero value otherwise.

### GetCustomMetadataOk

`func (o *ChainInfoResponse) GetCustomMetadataOk() (*string, bool)`

GetCustomMetadataOk returns a tuple with the CustomMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomMetadata

`func (o *ChainInfoResponse) SetCustomMetadata(v string)`

SetCustomMetadata sets CustomMetadata field to given value.

### HasCustomMetadata

`func (o *ChainInfoResponse) HasCustomMetadata() bool`

HasCustomMetadata returns a boolean if a field has been set.

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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


