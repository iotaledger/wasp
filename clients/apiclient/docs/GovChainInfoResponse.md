# GovChainInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainID** | **string** | ChainID (Bech32-encoded). | 
**ChainOwnerId** | **string** | The chain owner address (Bech32-encoded). | 
**EvmJsonRpcUrl** | **string** | The EVM json rpc url | 
**EvmWebSocketUrl** | **string** | The EVM websocket url | 
**GasFeePolicy** | [**FeePolicy**](FeePolicy.md) |  | 
**GasLimits** | [**Limits**](Limits.md) |  | 
**PublicUrl** | **string** | The fully qualified public url leading to the chains metadata | 

## Methods

### NewGovChainInfoResponse

`func NewGovChainInfoResponse(chainID string, chainOwnerId string, evmJsonRpcUrl string, evmWebSocketUrl string, gasFeePolicy FeePolicy, gasLimits Limits, publicUrl string, ) *GovChainInfoResponse`

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


### GetEvmJsonRpcUrl

`func (o *GovChainInfoResponse) GetEvmJsonRpcUrl() string`

GetEvmJsonRpcUrl returns the EvmJsonRpcUrl field if non-nil, zero value otherwise.

### GetEvmJsonRpcUrlOk

`func (o *GovChainInfoResponse) GetEvmJsonRpcUrlOk() (*string, bool)`

GetEvmJsonRpcUrlOk returns a tuple with the EvmJsonRpcUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmJsonRpcUrl

`func (o *GovChainInfoResponse) SetEvmJsonRpcUrl(v string)`

SetEvmJsonRpcUrl sets EvmJsonRpcUrl field to given value.


### GetEvmWebSocketUrl

`func (o *GovChainInfoResponse) GetEvmWebSocketUrl() string`

GetEvmWebSocketUrl returns the EvmWebSocketUrl field if non-nil, zero value otherwise.

### GetEvmWebSocketUrlOk

`func (o *GovChainInfoResponse) GetEvmWebSocketUrlOk() (*string, bool)`

GetEvmWebSocketUrlOk returns a tuple with the EvmWebSocketUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmWebSocketUrl

`func (o *GovChainInfoResponse) SetEvmWebSocketUrl(v string)`

SetEvmWebSocketUrl sets EvmWebSocketUrl field to given value.


### GetGasFeePolicy

`func (o *GovChainInfoResponse) GetGasFeePolicy() FeePolicy`

GetGasFeePolicy returns the GasFeePolicy field if non-nil, zero value otherwise.

### GetGasFeePolicyOk

`func (o *GovChainInfoResponse) GetGasFeePolicyOk() (*FeePolicy, bool)`

GetGasFeePolicyOk returns a tuple with the GasFeePolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeePolicy

`func (o *GovChainInfoResponse) SetGasFeePolicy(v FeePolicy)`

SetGasFeePolicy sets GasFeePolicy field to given value.


### GetGasLimits

`func (o *GovChainInfoResponse) GetGasLimits() Limits`

GetGasLimits returns the GasLimits field if non-nil, zero value otherwise.

### GetGasLimitsOk

`func (o *GovChainInfoResponse) GetGasLimitsOk() (*Limits, bool)`

GetGasLimitsOk returns a tuple with the GasLimits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasLimits

`func (o *GovChainInfoResponse) SetGasLimits(v Limits)`

SetGasLimits sets GasLimits field to given value.


### GetPublicUrl

`func (o *GovChainInfoResponse) GetPublicUrl() string`

GetPublicUrl returns the PublicUrl field if non-nil, zero value otherwise.

### GetPublicUrlOk

`func (o *GovChainInfoResponse) GetPublicUrlOk() (*string, bool)`

GetPublicUrlOk returns a tuple with the PublicUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicUrl

`func (o *GovChainInfoResponse) SetPublicUrl(v string)`

SetPublicUrl sets PublicUrl field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


