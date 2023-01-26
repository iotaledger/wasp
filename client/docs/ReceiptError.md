# ReceiptError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContractId** | Pointer to **int32** |  | [optional] 
**ErrorCode** | Pointer to **string** |  | [optional] 
**ErrorId** | Pointer to **int32** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**MessageFormat** | Pointer to **string** |  | [optional] 
**Parameters** | Pointer to **[]string** |  | [optional] 

## Methods

### NewReceiptError

`func NewReceiptError() *ReceiptError`

NewReceiptError instantiates a new ReceiptError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReceiptErrorWithDefaults

`func NewReceiptErrorWithDefaults() *ReceiptError`

NewReceiptErrorWithDefaults instantiates a new ReceiptError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContractId

`func (o *ReceiptError) GetContractId() int32`

GetContractId returns the ContractId field if non-nil, zero value otherwise.

### GetContractIdOk

`func (o *ReceiptError) GetContractIdOk() (*int32, bool)`

GetContractIdOk returns a tuple with the ContractId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractId

`func (o *ReceiptError) SetContractId(v int32)`

SetContractId sets ContractId field to given value.

### HasContractId

`func (o *ReceiptError) HasContractId() bool`

HasContractId returns a boolean if a field has been set.

### GetErrorCode

`func (o *ReceiptError) GetErrorCode() string`

GetErrorCode returns the ErrorCode field if non-nil, zero value otherwise.

### GetErrorCodeOk

`func (o *ReceiptError) GetErrorCodeOk() (*string, bool)`

GetErrorCodeOk returns a tuple with the ErrorCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorCode

`func (o *ReceiptError) SetErrorCode(v string)`

SetErrorCode sets ErrorCode field to given value.

### HasErrorCode

`func (o *ReceiptError) HasErrorCode() bool`

HasErrorCode returns a boolean if a field has been set.

### GetErrorId

`func (o *ReceiptError) GetErrorId() int32`

GetErrorId returns the ErrorId field if non-nil, zero value otherwise.

### GetErrorIdOk

`func (o *ReceiptError) GetErrorIdOk() (*int32, bool)`

GetErrorIdOk returns a tuple with the ErrorId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorId

`func (o *ReceiptError) SetErrorId(v int32)`

SetErrorId sets ErrorId field to given value.

### HasErrorId

`func (o *ReceiptError) HasErrorId() bool`

HasErrorId returns a boolean if a field has been set.

### GetMessage

`func (o *ReceiptError) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ReceiptError) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ReceiptError) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ReceiptError) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetMessageFormat

`func (o *ReceiptError) GetMessageFormat() string`

GetMessageFormat returns the MessageFormat field if non-nil, zero value otherwise.

### GetMessageFormatOk

`func (o *ReceiptError) GetMessageFormatOk() (*string, bool)`

GetMessageFormatOk returns a tuple with the MessageFormat field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessageFormat

`func (o *ReceiptError) SetMessageFormat(v string)`

SetMessageFormat sets MessageFormat field to given value.

### HasMessageFormat

`func (o *ReceiptError) HasMessageFormat() bool`

HasMessageFormat returns a boolean if a field has been set.

### GetParameters

`func (o *ReceiptError) GetParameters() []string`

GetParameters returns the Parameters field if non-nil, zero value otherwise.

### GetParametersOk

`func (o *ReceiptError) GetParametersOk() (*[]string, bool)`

GetParametersOk returns a tuple with the Parameters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParameters

`func (o *ReceiptError) SetParameters(v []string)`

SetParameters sets Parameters field to given value.

### HasParameters

`func (o *ReceiptError) HasParameters() bool`

HasParameters returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


