# BaseToken

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Decimals** | Pointer to **int32** | The token decimals | [optional] 
**Name** | Pointer to **string** | The base token name | [optional] 
**Subunit** | Pointer to **string** | The token subunit | [optional] 
**TickerSymbol** | Pointer to **string** | The ticker symbol | [optional] 
**Unit** | Pointer to **string** | The token unit | [optional] 
**UseMetricPrefix** | Pointer to **bool** | Whether or not the token uses a metric prefix | [optional] 

## Methods

### NewBaseToken

`func NewBaseToken() *BaseToken`

NewBaseToken instantiates a new BaseToken object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBaseTokenWithDefaults

`func NewBaseTokenWithDefaults() *BaseToken`

NewBaseTokenWithDefaults instantiates a new BaseToken object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDecimals

`func (o *BaseToken) GetDecimals() int32`

GetDecimals returns the Decimals field if non-nil, zero value otherwise.

### GetDecimalsOk

`func (o *BaseToken) GetDecimalsOk() (*int32, bool)`

GetDecimalsOk returns a tuple with the Decimals field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDecimals

`func (o *BaseToken) SetDecimals(v int32)`

SetDecimals sets Decimals field to given value.

### HasDecimals

`func (o *BaseToken) HasDecimals() bool`

HasDecimals returns a boolean if a field has been set.

### GetName

`func (o *BaseToken) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *BaseToken) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *BaseToken) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *BaseToken) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSubunit

`func (o *BaseToken) GetSubunit() string`

GetSubunit returns the Subunit field if non-nil, zero value otherwise.

### GetSubunitOk

`func (o *BaseToken) GetSubunitOk() (*string, bool)`

GetSubunitOk returns a tuple with the Subunit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubunit

`func (o *BaseToken) SetSubunit(v string)`

SetSubunit sets Subunit field to given value.

### HasSubunit

`func (o *BaseToken) HasSubunit() bool`

HasSubunit returns a boolean if a field has been set.

### GetTickerSymbol

`func (o *BaseToken) GetTickerSymbol() string`

GetTickerSymbol returns the TickerSymbol field if non-nil, zero value otherwise.

### GetTickerSymbolOk

`func (o *BaseToken) GetTickerSymbolOk() (*string, bool)`

GetTickerSymbolOk returns a tuple with the TickerSymbol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTickerSymbol

`func (o *BaseToken) SetTickerSymbol(v string)`

SetTickerSymbol sets TickerSymbol field to given value.

### HasTickerSymbol

`func (o *BaseToken) HasTickerSymbol() bool`

HasTickerSymbol returns a boolean if a field has been set.

### GetUnit

`func (o *BaseToken) GetUnit() string`

GetUnit returns the Unit field if non-nil, zero value otherwise.

### GetUnitOk

`func (o *BaseToken) GetUnitOk() (*string, bool)`

GetUnitOk returns a tuple with the Unit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnit

`func (o *BaseToken) SetUnit(v string)`

SetUnit sets Unit field to given value.

### HasUnit

`func (o *BaseToken) HasUnit() bool`

HasUnit returns a boolean if a field has been set.

### GetUseMetricPrefix

`func (o *BaseToken) GetUseMetricPrefix() bool`

GetUseMetricPrefix returns the UseMetricPrefix field if non-nil, zero value otherwise.

### GetUseMetricPrefixOk

`func (o *BaseToken) GetUseMetricPrefixOk() (*bool, bool)`

GetUseMetricPrefixOk returns a tuple with the UseMetricPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseMetricPrefix

`func (o *BaseToken) SetUseMetricPrefix(v bool)`

SetUseMetricPrefix sets UseMetricPrefix field to given value.

### HasUseMetricPrefix

`func (o *BaseToken) HasUseMetricPrefix() bool`

HasUseMetricPrefix returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


