# AssetsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseTokens** | Pointer to **int64** |  | [optional] 
**NativeTokens** | Pointer to [**[]NativeToken**](NativeToken.md) |  | [optional] 

## Methods

### NewAssetsResponse

`func NewAssetsResponse() *AssetsResponse`

NewAssetsResponse instantiates a new AssetsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsResponseWithDefaults

`func NewAssetsResponseWithDefaults() *AssetsResponse`

NewAssetsResponseWithDefaults instantiates a new AssetsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBaseTokens

`func (o *AssetsResponse) GetBaseTokens() int64`

GetBaseTokens returns the BaseTokens field if non-nil, zero value otherwise.

### GetBaseTokensOk

`func (o *AssetsResponse) GetBaseTokensOk() (*int64, bool)`

GetBaseTokensOk returns a tuple with the BaseTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseTokens

`func (o *AssetsResponse) SetBaseTokens(v int64)`

SetBaseTokens sets BaseTokens field to given value.

### HasBaseTokens

`func (o *AssetsResponse) HasBaseTokens() bool`

HasBaseTokens returns a boolean if a field has been set.

### GetNativeTokens

`func (o *AssetsResponse) GetNativeTokens() []NativeToken`

GetNativeTokens returns the NativeTokens field if non-nil, zero value otherwise.

### GetNativeTokensOk

`func (o *AssetsResponse) GetNativeTokensOk() (*[]NativeToken, bool)`

GetNativeTokensOk returns a tuple with the NativeTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNativeTokens

`func (o *AssetsResponse) SetNativeTokens(v []NativeToken)`

SetNativeTokens sets NativeTokens field to given value.

### HasNativeTokens

`func (o *AssetsResponse) HasNativeTokens() bool`

HasNativeTokens returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


