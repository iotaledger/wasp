# ChainInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainID** | **string** | ChainID (Bech32-encoded). | 
**ChainOwnerId** | **string** | The chain owner address (Bech32-encoded). | 
**EvmChainId** | **uint32** | The EVM chain ID | 
**EvmJsonRpcUrl** | **string** | The EVM json rpc url | 
**EvmWebSocketUrl** | **string** | The EVM websocket url | 
**GasFeePolicy** | [**FeePolicy**](FeePolicy.md) |  | 
**GasLimits** | [**Limits**](Limits.md) |  | 
**IsActive** | **bool** | Whether or not the chain is active. | 
**PublicUrl** | **string** | The fully qualified public url leading to the chains metadata | 
**Standard** | **string** | The chain info standard | 

## Methods

### NewChainInfoResponse

`func NewChainInfoResponse(chainID string, chainOwnerId string, evmChainId uint32, evmJsonRpcUrl string, evmWebSocketUrl string, gasFeePolicy FeePolicy, gasLimits Limits, isActive bool, publicUrl string, standard string, ) *ChainInfoResponse`

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


### GetEvmJsonRpcUrl

`func (o *ChainInfoResponse) GetEvmJsonRpcUrl() string`

GetEvmJsonRpcUrl returns the EvmJsonRpcUrl field if non-nil, zero value otherwise.

### GetEvmJsonRpcUrlOk

`func (o *ChainInfoResponse) GetEvmJsonRpcUrlOk() (*string, bool)`

GetEvmJsonRpcUrlOk returns a tuple with the EvmJsonRpcUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmJsonRpcUrl

`func (o *ChainInfoResponse) SetEvmJsonRpcUrl(v string)`

SetEvmJsonRpcUrl sets EvmJsonRpcUrl field to given value.


### GetEvmWebSocketUrl

`func (o *ChainInfoResponse) GetEvmWebSocketUrl() string`

GetEvmWebSocketUrl returns the EvmWebSocketUrl field if non-nil, zero value otherwise.

### GetEvmWebSocketUrlOk

`func (o *ChainInfoResponse) GetEvmWebSocketUrlOk() (*string, bool)`

GetEvmWebSocketUrlOk returns a tuple with the EvmWebSocketUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmWebSocketUrl

`func (o *ChainInfoResponse) SetEvmWebSocketUrl(v string)`

SetEvmWebSocketUrl sets EvmWebSocketUrl field to given value.


### GetGasFeePolicy

`func (o *ChainInfoResponse) GetGasFeePolicy() FeePolicy`

GetGasFeePolicy returns the GasFeePolicy field if non-nil, zero value otherwise.

### GetGasFeePolicyOk

`func (o *ChainInfoResponse) GetGasFeePolicyOk() (*FeePolicy, bool)`

GetGasFeePolicyOk returns a tuple with the GasFeePolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeePolicy

`func (o *ChainInfoResponse) SetGasFeePolicy(v FeePolicy)`

SetGasFeePolicy sets GasFeePolicy field to given value.


### GetGasLimits

`func (o *ChainInfoResponse) GetGasLimits() Limits`

GetGasLimits returns the GasLimits field if non-nil, zero value otherwise.

### GetGasLimitsOk

`func (o *ChainInfoResponse) GetGasLimitsOk() (*Limits, bool)`

GetGasLimitsOk returns a tuple with the GasLimits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasLimits

`func (o *ChainInfoResponse) SetGasLimits(v Limits)`

SetGasLimits sets GasLimits field to given value.


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


### GetPublicUrl

`func (o *ChainInfoResponse) GetPublicUrl() string`

GetPublicUrl returns the PublicUrl field if non-nil, zero value otherwise.

### GetPublicUrlOk

`func (o *ChainInfoResponse) GetPublicUrlOk() (*string, bool)`

GetPublicUrlOk returns a tuple with the PublicUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicUrl

`func (o *ChainInfoResponse) SetPublicUrl(v string)`

SetPublicUrl sets PublicUrl field to given value.


### GetStandard

`func (o *ChainInfoResponse) GetStandard() string`

GetStandard returns the Standard field if non-nil, zero value otherwise.

### GetStandardOk

`func (o *ChainInfoResponse) GetStandardOk() (*string, bool)`

GetStandardOk returns a tuple with the Standard field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStandard

`func (o *ChainInfoResponse) SetStandard(v string)`

SetStandard sets Standard field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


