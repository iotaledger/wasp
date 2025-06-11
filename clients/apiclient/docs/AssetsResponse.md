# AssetsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseTokens** | **string** | The base tokens (uint64 as string) | 
**Coins** | [**[]CoinJSON**](CoinJSON.md) |  | 
**Objects** | [**[]IotaObjectJSON**](IotaObjectJSON.md) |  | 

## Methods

### NewAssetsResponse

`func NewAssetsResponse(baseTokens string, coins []CoinJSON, objects []IotaObjectJSON, ) *AssetsResponse`

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

`func (o *AssetsResponse) GetBaseTokens() string`

GetBaseTokens returns the BaseTokens field if non-nil, zero value otherwise.

### GetBaseTokensOk

`func (o *AssetsResponse) GetBaseTokensOk() (*string, bool)`

GetBaseTokensOk returns a tuple with the BaseTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseTokens

`func (o *AssetsResponse) SetBaseTokens(v string)`

SetBaseTokens sets BaseTokens field to given value.


### GetCoins

`func (o *AssetsResponse) GetCoins() []CoinJSON`

GetCoins returns the Coins field if non-nil, zero value otherwise.

### GetCoinsOk

`func (o *AssetsResponse) GetCoinsOk() (*[]CoinJSON, bool)`

GetCoinsOk returns a tuple with the Coins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCoins

`func (o *AssetsResponse) SetCoins(v []CoinJSON)`

SetCoins sets Coins field to given value.


### GetObjects

`func (o *AssetsResponse) GetObjects() []IotaObjectJSON`

GetObjects returns the Objects field if non-nil, zero value otherwise.

### GetObjectsOk

`func (o *AssetsResponse) GetObjectsOk() (*[]IotaObjectJSON, bool)`

GetObjectsOk returns a tuple with the Objects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjects

`func (o *AssetsResponse) SetObjects(v []IotaObjectJSON)`

SetObjects sets Objects field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


