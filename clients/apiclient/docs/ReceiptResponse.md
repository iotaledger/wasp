# ReceiptResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BlockIndex** | **uint32** |  | 
**ErrorMessage** | Pointer to **string** |  | [optional] 
**GasBudget** | **string** | The gas budget (uint64 as string) | 
**GasBurnLog** | [**[]BurnRecord**](BurnRecord.md) |  | 
**GasBurned** | **string** | The burned gas (uint64 as string) | 
**GasFeeCharged** | **string** | The charged gas fee (uint64 as string) | 
**RawError** | Pointer to [**UnresolvedVMErrorJSON**](UnresolvedVMErrorJSON.md) |  | [optional] 
**Request** | [**RequestJSON**](RequestJSON.md) |  | 
**RequestIndex** | **uint32** |  | 
**StorageDepositCharged** | **string** | Storage deposit charged (uint64 as string) | 

## Methods

### NewReceiptResponse

`func NewReceiptResponse(blockIndex uint32, gasBudget string, gasBurnLog []BurnRecord, gasBurned string, gasFeeCharged string, request RequestJSON, requestIndex uint32, storageDepositCharged string, ) *ReceiptResponse`

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


### GetErrorMessage

`func (o *ReceiptResponse) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *ReceiptResponse) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *ReceiptResponse) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.

### HasErrorMessage

`func (o *ReceiptResponse) HasErrorMessage() bool`

HasErrorMessage returns a boolean if a field has been set.

### GetGasBudget

`func (o *ReceiptResponse) GetGasBudget() string`

GetGasBudget returns the GasBudget field if non-nil, zero value otherwise.

### GetGasBudgetOk

`func (o *ReceiptResponse) GetGasBudgetOk() (*string, bool)`

GetGasBudgetOk returns a tuple with the GasBudget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBudget

`func (o *ReceiptResponse) SetGasBudget(v string)`

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

`func (o *ReceiptResponse) GetGasBurned() string`

GetGasBurned returns the GasBurned field if non-nil, zero value otherwise.

### GetGasBurnedOk

`func (o *ReceiptResponse) GetGasBurnedOk() (*string, bool)`

GetGasBurnedOk returns a tuple with the GasBurned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBurned

`func (o *ReceiptResponse) SetGasBurned(v string)`

SetGasBurned sets GasBurned field to given value.


### GetGasFeeCharged

`func (o *ReceiptResponse) GetGasFeeCharged() string`

GetGasFeeCharged returns the GasFeeCharged field if non-nil, zero value otherwise.

### GetGasFeeChargedOk

`func (o *ReceiptResponse) GetGasFeeChargedOk() (*string, bool)`

GetGasFeeChargedOk returns a tuple with the GasFeeCharged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeeCharged

`func (o *ReceiptResponse) SetGasFeeCharged(v string)`

SetGasFeeCharged sets GasFeeCharged field to given value.


### GetRawError

`func (o *ReceiptResponse) GetRawError() UnresolvedVMErrorJSON`

GetRawError returns the RawError field if non-nil, zero value otherwise.

### GetRawErrorOk

`func (o *ReceiptResponse) GetRawErrorOk() (*UnresolvedVMErrorJSON, bool)`

GetRawErrorOk returns a tuple with the RawError field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRawError

`func (o *ReceiptResponse) SetRawError(v UnresolvedVMErrorJSON)`

SetRawError sets RawError field to given value.

### HasRawError

`func (o *ReceiptResponse) HasRawError() bool`

HasRawError returns a boolean if a field has been set.

### GetRequest

`func (o *ReceiptResponse) GetRequest() RequestJSON`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *ReceiptResponse) GetRequestOk() (*RequestJSON, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *ReceiptResponse) SetRequest(v RequestJSON)`

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


### GetStorageDepositCharged

`func (o *ReceiptResponse) GetStorageDepositCharged() string`

GetStorageDepositCharged returns the StorageDepositCharged field if non-nil, zero value otherwise.

### GetStorageDepositChargedOk

`func (o *ReceiptResponse) GetStorageDepositChargedOk() (*string, bool)`

GetStorageDepositChargedOk returns a tuple with the StorageDepositCharged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageDepositCharged

`func (o *ReceiptResponse) SetStorageDepositCharged(v string)`

SetStorageDepositCharged sets StorageDepositCharged field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


