# CoinJSON

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Balance** | [**BigInt**](BigInt.md) |  | 
**CoinType** | [**Type**](Type.md) |  | 

## Methods

### NewCoinJSON

`func NewCoinJSON(balance BigInt, coinType Type, ) *CoinJSON`

NewCoinJSON instantiates a new CoinJSON object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCoinJSONWithDefaults

`func NewCoinJSONWithDefaults() *CoinJSON`

NewCoinJSONWithDefaults instantiates a new CoinJSON object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBalance

`func (o *CoinJSON) GetBalance() BigInt`

GetBalance returns the Balance field if non-nil, zero value otherwise.

### GetBalanceOk

`func (o *CoinJSON) GetBalanceOk() (*BigInt, bool)`

GetBalanceOk returns a tuple with the Balance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBalance

`func (o *CoinJSON) SetBalance(v BigInt)`

SetBalance sets Balance field to given value.


### GetCoinType

`func (o *CoinJSON) GetCoinType() Type`

GetCoinType returns the CoinType field if non-nil, zero value otherwise.

### GetCoinTypeOk

`func (o *CoinJSON) GetCoinTypeOk() (*Type, bool)`

GetCoinTypeOk returns a tuple with the CoinType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCoinType

`func (o *CoinJSON) SetCoinType(v Type)`

SetCoinType sets CoinType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


