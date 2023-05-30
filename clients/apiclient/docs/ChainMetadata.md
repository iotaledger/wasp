# ChainMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | **string** | The description of the chain. | 
**EvmJsonRpcUrl** | **string** | The EVM json rpc url | 
**EvmWebSocketUrl** | **string** | The EVM websocket url) | 
**Name** | **string** | The name of the chain | 
**Website** | **string** | The official website of the chain. | 

## Methods

### NewChainMetadata

`func NewChainMetadata(description string, evmJsonRpcUrl string, evmWebSocketUrl string, name string, website string, ) *ChainMetadata`

NewChainMetadata instantiates a new ChainMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewChainMetadataWithDefaults

`func NewChainMetadataWithDefaults() *ChainMetadata`

NewChainMetadataWithDefaults instantiates a new ChainMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *ChainMetadata) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ChainMetadata) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ChainMetadata) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetEvmJsonRpcUrl

`func (o *ChainMetadata) GetEvmJsonRpcUrl() string`

GetEvmJsonRpcUrl returns the EvmJsonRpcUrl field if non-nil, zero value otherwise.

### GetEvmJsonRpcUrlOk

`func (o *ChainMetadata) GetEvmJsonRpcUrlOk() (*string, bool)`

GetEvmJsonRpcUrlOk returns a tuple with the EvmJsonRpcUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmJsonRpcUrl

`func (o *ChainMetadata) SetEvmJsonRpcUrl(v string)`

SetEvmJsonRpcUrl sets EvmJsonRpcUrl field to given value.


### GetEvmWebSocketUrl

`func (o *ChainMetadata) GetEvmWebSocketUrl() string`

GetEvmWebSocketUrl returns the EvmWebSocketUrl field if non-nil, zero value otherwise.

### GetEvmWebSocketUrlOk

`func (o *ChainMetadata) GetEvmWebSocketUrlOk() (*string, bool)`

GetEvmWebSocketUrlOk returns a tuple with the EvmWebSocketUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmWebSocketUrl

`func (o *ChainMetadata) SetEvmWebSocketUrl(v string)`

SetEvmWebSocketUrl sets EvmWebSocketUrl field to given value.


### GetName

`func (o *ChainMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ChainMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ChainMetadata) SetName(v string)`

SetName sets Name field to given value.


### GetWebsite

`func (o *ChainMetadata) GetWebsite() string`

GetWebsite returns the Website field if non-nil, zero value otherwise.

### GetWebsiteOk

`func (o *ChainMetadata) GetWebsiteOk() (*string, bool)`

GetWebsiteOk returns a tuple with the Website field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebsite

`func (o *ChainMetadata) SetWebsite(v string)`

SetWebsite sets Website field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


