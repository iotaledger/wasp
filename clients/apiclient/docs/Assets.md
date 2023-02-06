# Assets

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseTokens** | **string** | The base tokens (uint64 as string) | 
**NativeTokens** | [**[]NativeToken**](NativeToken.md) |  | 
**Nfts** | **[]string** |  | 

## Methods

### NewAssets

`func NewAssets(baseTokens string, nativeTokens []NativeToken, nfts []string, ) *Assets`

NewAssets instantiates a new Assets object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsWithDefaults

`func NewAssetsWithDefaults() *Assets`

NewAssetsWithDefaults instantiates a new Assets object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBaseTokens

`func (o *Assets) GetBaseTokens() string`

GetBaseTokens returns the BaseTokens field if non-nil, zero value otherwise.

### GetBaseTokensOk

`func (o *Assets) GetBaseTokensOk() (*string, bool)`

GetBaseTokensOk returns a tuple with the BaseTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseTokens

`func (o *Assets) SetBaseTokens(v string)`

SetBaseTokens sets BaseTokens field to given value.


### GetNativeTokens

`func (o *Assets) GetNativeTokens() []NativeToken`

GetNativeTokens returns the NativeTokens field if non-nil, zero value otherwise.

### GetNativeTokensOk

`func (o *Assets) GetNativeTokensOk() (*[]NativeToken, bool)`

GetNativeTokensOk returns a tuple with the NativeTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNativeTokens

`func (o *Assets) SetNativeTokens(v []NativeToken)`

SetNativeTokens sets NativeTokens field to given value.


### GetNfts

`func (o *Assets) GetNfts() []string`

GetNfts returns the Nfts field if non-nil, zero value otherwise.

### GetNftsOk

`func (o *Assets) GetNftsOk() (*[]string, bool)`

GetNftsOk returns a tuple with the Nfts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfts

`func (o *Assets) SetNfts(v []string)`

SetNfts sets Nfts field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


