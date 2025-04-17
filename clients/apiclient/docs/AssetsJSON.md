# AssetsJSON

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Coins** | [**[]CoinJSON**](CoinJSON.md) |  | 
**Objects** | [**[]IotaObject**](IotaObject.md) |  | 

## Methods

### NewAssetsJSON

`func NewAssetsJSON(coins []CoinJSON, objects []IotaObject, ) *AssetsJSON`

NewAssetsJSON instantiates a new AssetsJSON object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsJSONWithDefaults

`func NewAssetsJSONWithDefaults() *AssetsJSON`

NewAssetsJSONWithDefaults instantiates a new AssetsJSON object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCoins

`func (o *AssetsJSON) GetCoins() []CoinJSON`

GetCoins returns the Coins field if non-nil, zero value otherwise.

### GetCoinsOk

`func (o *AssetsJSON) GetCoinsOk() (*[]CoinJSON, bool)`

GetCoinsOk returns a tuple with the Coins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCoins

`func (o *AssetsJSON) SetCoins(v []CoinJSON)`

SetCoins sets Coins field to given value.


### GetObjects

`func (o *AssetsJSON) GetObjects() []IotaObject`

GetObjects returns the Objects field if non-nil, zero value otherwise.

### GetObjectsOk

`func (o *AssetsJSON) GetObjectsOk() (*[]IotaObject, bool)`

GetObjectsOk returns a tuple with the Objects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjects

`func (o *AssetsJSON) SetObjects(v []IotaObject)`

SetObjects sets Objects field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


