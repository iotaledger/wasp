# BaseToken

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Decimals** | **int32** | The token decimals | 
**Name** | **string** | The base token name | 
**Subunit** | **string** | The token subunit | 
**TickerSymbol** | **string** | The ticker symbol | 
**Unit** | **string** | The token unit | 
**UseMetricPrefix** | **bool** | Whether or not the token uses a metric prefix | 

## Methods

### NewBaseToken

`func NewBaseToken(decimals int32, name string, subunit string, tickerSymbol string, unit string, useMetricPrefix bool, ) *BaseToken`

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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


