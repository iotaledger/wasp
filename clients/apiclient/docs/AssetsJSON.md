# AssetsJSON

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseTokens** | **string** | The base tokens (uint64 as string) | 
**NativeTokens** | [**[]NativeTokenJSON**](NativeTokenJSON.md) |  | 
**Nfts** | **[]string** |  | 

## Methods

### NewAssetsJSON

`func NewAssetsJSON(baseTokens string, nativeTokens []NativeTokenJSON, nfts []string, ) *AssetsJSON`

NewAssetsJSON instantiates a new AssetsJSON object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsJSONWithDefaults

`func NewAssetsJSONWithDefaults() *AssetsJSON`

NewAssetsJSONWithDefaults instantiates a new AssetsJSON object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBaseTokens

`func (o *AssetsJSON) GetBaseTokens() string`

GetBaseTokens returns the BaseTokens field if non-nil, zero value otherwise.

### GetBaseTokensOk

`func (o *AssetsJSON) GetBaseTokensOk() (*string, bool)`

GetBaseTokensOk returns a tuple with the BaseTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseTokens

`func (o *AssetsJSON) SetBaseTokens(v string)`

SetBaseTokens sets BaseTokens field to given value.


### GetNativeTokens

`func (o *AssetsJSON) GetNativeTokens() []NativeTokenJSON`

GetNativeTokens returns the NativeTokens field if non-nil, zero value otherwise.

### GetNativeTokensOk

`func (o *AssetsJSON) GetNativeTokensOk() (*[]NativeTokenJSON, bool)`

GetNativeTokensOk returns a tuple with the NativeTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNativeTokens

`func (o *AssetsJSON) SetNativeTokens(v []NativeTokenJSON)`

SetNativeTokens sets NativeTokens field to given value.


### GetNfts

`func (o *AssetsJSON) GetNfts() []string`

GetNfts returns the Nfts field if non-nil, zero value otherwise.

### GetNftsOk

`func (o *AssetsJSON) GetNftsOk() (*[]string, bool)`

GetNftsOk returns a tuple with the Nfts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfts

`func (o *AssetsJSON) SetNfts(v []string)`

SetNfts sets Nfts field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


