# RequestJSON

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Allowance** | [**AssetsJSON**](AssetsJSON.md) |  | 
**AllowanceError** | Pointer to **string** |  | [optional] 
**Assets** | [**AssetsJSON**](AssetsJSON.md) |  | 
**CallTarget** | [**CallTargetJSON**](CallTargetJSON.md) |  | 
**GasBudget** | **string** | The gas budget (uint64 as string) | 
**IsEVM** | **bool** |  | 
**IsOffLedger** | **bool** |  | 
**Params** | **[]string** |  | 
**RequestId** | **string** |  | 
**SenderAccount** | **string** |  | 

## Methods

### NewRequestJSON

`func NewRequestJSON(allowance AssetsJSON, assets AssetsJSON, callTarget CallTargetJSON, gasBudget string, isEVM bool, isOffLedger bool, params []string, requestId string, senderAccount string, ) *RequestJSON`

NewRequestJSON instantiates a new RequestJSON object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRequestJSONWithDefaults

`func NewRequestJSONWithDefaults() *RequestJSON`

NewRequestJSONWithDefaults instantiates a new RequestJSON object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowance

`func (o *RequestJSON) GetAllowance() AssetsJSON`

GetAllowance returns the Allowance field if non-nil, zero value otherwise.

### GetAllowanceOk

`func (o *RequestJSON) GetAllowanceOk() (*AssetsJSON, bool)`

GetAllowanceOk returns a tuple with the Allowance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowance

`func (o *RequestJSON) SetAllowance(v AssetsJSON)`

SetAllowance sets Allowance field to given value.


### GetAllowanceError

`func (o *RequestJSON) GetAllowanceError() string`

GetAllowanceError returns the AllowanceError field if non-nil, zero value otherwise.

### GetAllowanceErrorOk

`func (o *RequestJSON) GetAllowanceErrorOk() (*string, bool)`

GetAllowanceErrorOk returns a tuple with the AllowanceError field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowanceError

`func (o *RequestJSON) SetAllowanceError(v string)`

SetAllowanceError sets AllowanceError field to given value.

### HasAllowanceError

`func (o *RequestJSON) HasAllowanceError() bool`

HasAllowanceError returns a boolean if a field has been set.

### GetAssets

`func (o *RequestJSON) GetAssets() AssetsJSON`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *RequestJSON) GetAssetsOk() (*AssetsJSON, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *RequestJSON) SetAssets(v AssetsJSON)`

SetAssets sets Assets field to given value.


### GetCallTarget

`func (o *RequestJSON) GetCallTarget() CallTargetJSON`

GetCallTarget returns the CallTarget field if non-nil, zero value otherwise.

### GetCallTargetOk

`func (o *RequestJSON) GetCallTargetOk() (*CallTargetJSON, bool)`

GetCallTargetOk returns a tuple with the CallTarget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCallTarget

`func (o *RequestJSON) SetCallTarget(v CallTargetJSON)`

SetCallTarget sets CallTarget field to given value.


### GetGasBudget

`func (o *RequestJSON) GetGasBudget() string`

GetGasBudget returns the GasBudget field if non-nil, zero value otherwise.

### GetGasBudgetOk

`func (o *RequestJSON) GetGasBudgetOk() (*string, bool)`

GetGasBudgetOk returns a tuple with the GasBudget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBudget

`func (o *RequestJSON) SetGasBudget(v string)`

SetGasBudget sets GasBudget field to given value.


### GetIsEVM

`func (o *RequestJSON) GetIsEVM() bool`

GetIsEVM returns the IsEVM field if non-nil, zero value otherwise.

### GetIsEVMOk

`func (o *RequestJSON) GetIsEVMOk() (*bool, bool)`

GetIsEVMOk returns a tuple with the IsEVM field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsEVM

`func (o *RequestJSON) SetIsEVM(v bool)`

SetIsEVM sets IsEVM field to given value.


### GetIsOffLedger

`func (o *RequestJSON) GetIsOffLedger() bool`

GetIsOffLedger returns the IsOffLedger field if non-nil, zero value otherwise.

### GetIsOffLedgerOk

`func (o *RequestJSON) GetIsOffLedgerOk() (*bool, bool)`

GetIsOffLedgerOk returns a tuple with the IsOffLedger field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsOffLedger

`func (o *RequestJSON) SetIsOffLedger(v bool)`

SetIsOffLedger sets IsOffLedger field to given value.


### GetParams

`func (o *RequestJSON) GetParams() []string`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *RequestJSON) GetParamsOk() (*[]string, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *RequestJSON) SetParams(v []string)`

SetParams sets Params field to given value.


### GetRequestId

`func (o *RequestJSON) GetRequestId() string`

GetRequestId returns the RequestId field if non-nil, zero value otherwise.

### GetRequestIdOk

`func (o *RequestJSON) GetRequestIdOk() (*string, bool)`

GetRequestIdOk returns a tuple with the RequestId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestId

`func (o *RequestJSON) SetRequestId(v string)`

SetRequestId sets RequestId field to given value.


### GetSenderAccount

`func (o *RequestJSON) GetSenderAccount() string`

GetSenderAccount returns the SenderAccount field if non-nil, zero value otherwise.

### GetSenderAccountOk

`func (o *RequestJSON) GetSenderAccountOk() (*string, bool)`

GetSenderAccountOk returns a tuple with the SenderAccount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSenderAccount

`func (o *RequestJSON) SetSenderAccount(v string)`

SetSenderAccount sets SenderAccount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


