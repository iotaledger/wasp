# ReceiptResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BlockIndex** | **uint32** |  | 
**Error** | Pointer to [**ReceiptError**](ReceiptError.md) |  | [optional] 
**GasBudget** | **int64** |  | 
**GasBurnLog** | [**[]BurnRecord**](BurnRecord.md) |  | 
**GasBurned** | **int64** |  | 
**GasFeeCharged** | **int64** |  | 
**Request** | **string** |  | 
**RequestIndex** | **uint32** |  | 

## Methods

### NewReceiptResponse

`func NewReceiptResponse(blockIndex uint32, gasBudget int64, gasBurnLog []BurnRecord, gasBurned int64, gasFeeCharged int64, request string, requestIndex uint32, ) *ReceiptResponse`

NewReceiptResponse instantiates a new ReceiptResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReceiptResponseWithDefaults

`func NewReceiptResponseWithDefaults() *ReceiptResponse`

NewReceiptResponseWithDefaults instantiates a new ReceiptResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBlockIndex

`func (o *ReceiptResponse) GetBlockIndex() uint32`

GetBlockIndex returns the BlockIndex field if non-nil, zero value otherwise.

### GetBlockIndexOk

`func (o *ReceiptResponse) GetBlockIndexOk() (*uint32, bool)`

GetBlockIndexOk returns a tuple with the BlockIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlockIndex

`func (o *ReceiptResponse) SetBlockIndex(v uint32)`

SetBlockIndex sets BlockIndex field to given value.


### GetError

`func (o *ReceiptResponse) GetError() ReceiptError`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *ReceiptResponse) GetErrorOk() (*ReceiptError, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *ReceiptResponse) SetError(v ReceiptError)`

SetError sets Error field to given value.

### HasError

`func (o *ReceiptResponse) HasError() bool`

HasError returns a boolean if a field has been set.

### GetGasBudget

`func (o *ReceiptResponse) GetGasBudget() int64`

GetGasBudget returns the GasBudget field if non-nil, zero value otherwise.

### GetGasBudgetOk

`func (o *ReceiptResponse) GetGasBudgetOk() (*int64, bool)`

GetGasBudgetOk returns a tuple with the GasBudget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBudget

`func (o *ReceiptResponse) SetGasBudget(v int64)`

SetGasBudget sets GasBudget field to given value.


### GetGasBurnLog

`func (o *ReceiptResponse) GetGasBurnLog() []BurnRecord`

GetGasBurnLog returns the GasBurnLog field if non-nil, zero value otherwise.

### GetGasBurnLogOk

`func (o *ReceiptResponse) GetGasBurnLogOk() (*[]BurnRecord, bool)`

GetGasBurnLogOk returns a tuple with the GasBurnLog field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBurnLog

`func (o *ReceiptResponse) SetGasBurnLog(v []BurnRecord)`

SetGasBurnLog sets GasBurnLog field to given value.


### GetGasBurned

`func (o *ReceiptResponse) GetGasBurned() int64`

GetGasBurned returns the GasBurned field if non-nil, zero value otherwise.

### GetGasBurnedOk

`func (o *ReceiptResponse) GetGasBurnedOk() (*int64, bool)`

GetGasBurnedOk returns a tuple with the GasBurned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBurned

`func (o *ReceiptResponse) SetGasBurned(v int64)`

SetGasBurned sets GasBurned field to given value.


### GetGasFeeCharged

`func (o *ReceiptResponse) GetGasFeeCharged() int64`

GetGasFeeCharged returns the GasFeeCharged field if non-nil, zero value otherwise.

### GetGasFeeChargedOk

`func (o *ReceiptResponse) GetGasFeeChargedOk() (*int64, bool)`

GetGasFeeChargedOk returns a tuple with the GasFeeCharged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeeCharged

`func (o *ReceiptResponse) SetGasFeeCharged(v int64)`

SetGasFeeCharged sets GasFeeCharged field to given value.


### GetRequest

`func (o *ReceiptResponse) GetRequest() string`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *ReceiptResponse) GetRequestOk() (*string, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *ReceiptResponse) SetRequest(v string)`

SetRequest sets Request field to given value.


### GetRequestIndex

`func (o *ReceiptResponse) GetRequestIndex() uint32`

GetRequestIndex returns the RequestIndex field if non-nil, zero value otherwise.

### GetRequestIndexOk

`func (o *ReceiptResponse) GetRequestIndexOk() (*uint32, bool)`

GetRequestIndexOk returns a tuple with the RequestIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestIndex

`func (o *ReceiptResponse) SetRequestIndex(v uint32)`

SetRequestIndex sets RequestIndex field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


