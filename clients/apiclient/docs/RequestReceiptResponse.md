# RequestReceiptResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BlockIndex** | **uint32** |  | 
**Error** | Pointer to [**BlockReceiptError**](BlockReceiptError.md) |  | [optional] 
**GasBudget** | **string** | The gas budget (uint64 as string) | 
**GasBurnLog** | [**BurnLog**](BurnLog.md) |  | 
**GasBurned** | **string** | The burned gas (uint64 as string) | 
**GasFeeCharged** | **string** | The charged gas fee (uint64 as string) | 
**Request** | [**RequestDetail**](RequestDetail.md) |  | 
**RequestIndex** | **uint32** |  | 

## Methods

### NewRequestReceiptResponse

`func NewRequestReceiptResponse(blockIndex uint32, gasBudget string, gasBurnLog BurnLog, gasBurned string, gasFeeCharged string, request RequestDetail, requestIndex uint32, ) *RequestReceiptResponse`

NewRequestReceiptResponse instantiates a new RequestReceiptResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRequestReceiptResponseWithDefaults

`func NewRequestReceiptResponseWithDefaults() *RequestReceiptResponse`

NewRequestReceiptResponseWithDefaults instantiates a new RequestReceiptResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBlockIndex

`func (o *RequestReceiptResponse) GetBlockIndex() uint32`

GetBlockIndex returns the BlockIndex field if non-nil, zero value otherwise.

### GetBlockIndexOk

`func (o *RequestReceiptResponse) GetBlockIndexOk() (*uint32, bool)`

GetBlockIndexOk returns a tuple with the BlockIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlockIndex

`func (o *RequestReceiptResponse) SetBlockIndex(v uint32)`

SetBlockIndex sets BlockIndex field to given value.


### GetError

`func (o *RequestReceiptResponse) GetError() BlockReceiptError`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *RequestReceiptResponse) GetErrorOk() (*BlockReceiptError, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *RequestReceiptResponse) SetError(v BlockReceiptError)`

SetError sets Error field to given value.

### HasError

`func (o *RequestReceiptResponse) HasError() bool`

HasError returns a boolean if a field has been set.

### GetGasBudget

`func (o *RequestReceiptResponse) GetGasBudget() string`

GetGasBudget returns the GasBudget field if non-nil, zero value otherwise.

### GetGasBudgetOk

`func (o *RequestReceiptResponse) GetGasBudgetOk() (*string, bool)`

GetGasBudgetOk returns a tuple with the GasBudget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBudget

`func (o *RequestReceiptResponse) SetGasBudget(v string)`

SetGasBudget sets GasBudget field to given value.


### GetGasBurnLog

`func (o *RequestReceiptResponse) GetGasBurnLog() BurnLog`

GetGasBurnLog returns the GasBurnLog field if non-nil, zero value otherwise.

### GetGasBurnLogOk

`func (o *RequestReceiptResponse) GetGasBurnLogOk() (*BurnLog, bool)`

GetGasBurnLogOk returns a tuple with the GasBurnLog field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBurnLog

`func (o *RequestReceiptResponse) SetGasBurnLog(v BurnLog)`

SetGasBurnLog sets GasBurnLog field to given value.


### GetGasBurned

`func (o *RequestReceiptResponse) GetGasBurned() string`

GetGasBurned returns the GasBurned field if non-nil, zero value otherwise.

### GetGasBurnedOk

`func (o *RequestReceiptResponse) GetGasBurnedOk() (*string, bool)`

GetGasBurnedOk returns a tuple with the GasBurned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBurned

`func (o *RequestReceiptResponse) SetGasBurned(v string)`

SetGasBurned sets GasBurned field to given value.


### GetGasFeeCharged

`func (o *RequestReceiptResponse) GetGasFeeCharged() string`

GetGasFeeCharged returns the GasFeeCharged field if non-nil, zero value otherwise.

### GetGasFeeChargedOk

`func (o *RequestReceiptResponse) GetGasFeeChargedOk() (*string, bool)`

GetGasFeeChargedOk returns a tuple with the GasFeeCharged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeeCharged

`func (o *RequestReceiptResponse) SetGasFeeCharged(v string)`

SetGasFeeCharged sets GasFeeCharged field to given value.


### GetRequest

`func (o *RequestReceiptResponse) GetRequest() RequestDetail`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *RequestReceiptResponse) GetRequestOk() (*RequestDetail, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *RequestReceiptResponse) SetRequest(v RequestDetail)`

SetRequest sets Request field to given value.


### GetRequestIndex

`func (o *RequestReceiptResponse) GetRequestIndex() uint32`

GetRequestIndex returns the RequestIndex field if non-nil, zero value otherwise.

### GetRequestIndexOk

`func (o *RequestReceiptResponse) GetRequestIndexOk() (*uint32, bool)`

GetRequestIndexOk returns a tuple with the RequestIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestIndex

`func (o *RequestReceiptResponse) SetRequestIndex(v uint32)`

SetRequestIndex sets RequestIndex field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


