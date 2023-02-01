# FungibleTokens

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseTokens** | **int64** |  | 
**NativeTokens** | [**[]NativeToken**](NativeToken.md) |  | 

## Methods

### NewFungibleTokens

`func NewFungibleTokens(baseTokens int64, nativeTokens []NativeToken, ) *FungibleTokens`

NewFungibleTokens instantiates a new FungibleTokens object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFungibleTokensWithDefaults

`func NewFungibleTokensWithDefaults() *FungibleTokens`

NewFungibleTokensWithDefaults instantiates a new FungibleTokens object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBaseTokens

`func (o *FungibleTokens) GetBaseTokens() int64`

GetBaseTokens returns the BaseTokens field if non-nil, zero value otherwise.

### GetBaseTokensOk

`func (o *FungibleTokens) GetBaseTokensOk() (*int64, bool)`

GetBaseTokensOk returns a tuple with the BaseTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseTokens

`func (o *FungibleTokens) SetBaseTokens(v int64)`

SetBaseTokens sets BaseTokens field to given value.


### GetNativeTokens

`func (o *FungibleTokens) GetNativeTokens() []NativeToken`

GetNativeTokens returns the NativeTokens field if non-nil, zero value otherwise.

### GetNativeTokensOk

`func (o *FungibleTokens) GetNativeTokensOk() (*[]NativeToken, bool)`

GetNativeTokensOk returns a tuple with the NativeTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNativeTokens

`func (o *FungibleTokens) SetNativeTokens(v []NativeToken)`

SetNativeTokens sets NativeTokens field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


