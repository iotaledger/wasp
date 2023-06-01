# PublicChainMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | **string** | The description of the chain. | 
**EvmJsonRpcURL** | **string** | The EVM json rpc url | 
**EvmWebSocketURL** | **string** | The EVM websocket url) | 
**Name** | **string** | The name of the chain | 
**Website** | **string** | The official website of the chain. | 

## Methods

### NewPublicChainMetadata

`func NewPublicChainMetadata(description string, evmJsonRpcURL string, evmWebSocketURL string, name string, website string, ) *PublicChainMetadata`

NewPublicChainMetadata instantiates a new PublicChainMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPublicChainMetadataWithDefaults

`func NewPublicChainMetadataWithDefaults() *PublicChainMetadata`

NewPublicChainMetadataWithDefaults instantiates a new PublicChainMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *PublicChainMetadata) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *PublicChainMetadata) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *PublicChainMetadata) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetEvmJsonRpcURL

`func (o *PublicChainMetadata) GetEvmJsonRpcURL() string`

GetEvmJsonRpcURL returns the EvmJsonRpcURL field if non-nil, zero value otherwise.

### GetEvmJsonRpcURLOk

`func (o *PublicChainMetadata) GetEvmJsonRpcURLOk() (*string, bool)`

GetEvmJsonRpcURLOk returns a tuple with the EvmJsonRpcURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmJsonRpcURL

`func (o *PublicChainMetadata) SetEvmJsonRpcURL(v string)`

SetEvmJsonRpcURL sets EvmJsonRpcURL field to given value.


### GetEvmWebSocketURL

`func (o *PublicChainMetadata) GetEvmWebSocketURL() string`

GetEvmWebSocketURL returns the EvmWebSocketURL field if non-nil, zero value otherwise.

### GetEvmWebSocketURLOk

`func (o *PublicChainMetadata) GetEvmWebSocketURLOk() (*string, bool)`

GetEvmWebSocketURLOk returns a tuple with the EvmWebSocketURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvmWebSocketURL

`func (o *PublicChainMetadata) SetEvmWebSocketURL(v string)`

SetEvmWebSocketURL sets EvmWebSocketURL field to given value.


### GetName

`func (o *PublicChainMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PublicChainMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PublicChainMetadata) SetName(v string)`

SetName sets Name field to given value.


### GetWebsite

`func (o *PublicChainMetadata) GetWebsite() string`

GetWebsite returns the Website field if non-nil, zero value otherwise.

### GetWebsiteOk

`func (o *PublicChainMetadata) GetWebsiteOk() (*string, bool)`

GetWebsiteOk returns a tuple with the Website field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebsite

`func (o *PublicChainMetadata) SetWebsite(v string)`

SetWebsite sets Website field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


