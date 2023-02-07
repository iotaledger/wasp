# ReceiptError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContractHName** | **string** | The contract hname (Hex) | 
**ErrorCode** | **string** |  | 
**ErrorId** | **uint32** |  | 
**Message** | **string** |  | 
**MessageFormat** | **string** |  | 
**Parameters** | **[]string** |  | 

## Methods

### NewReceiptError

`func NewReceiptError(contractHName string, errorCode string, errorId uint32, message string, messageFormat string, parameters []string, ) *ReceiptError`

NewReceiptError instantiates a new ReceiptError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReceiptErrorWithDefaults

`func NewReceiptErrorWithDefaults() *ReceiptError`

NewReceiptErrorWithDefaults instantiates a new ReceiptError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContractHName

`func (o *ReceiptError) GetContractHName() string`

GetContractHName returns the ContractHName field if non-nil, zero value otherwise.

### GetContractHNameOk

`func (o *ReceiptError) GetContractHNameOk() (*string, bool)`

GetContractHNameOk returns a tuple with the ContractHName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractHName

`func (o *ReceiptError) SetContractHName(v string)`

SetContractHName sets ContractHName field to given value.


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


### GetErrorId

`func (o *ReceiptError) GetErrorId() uint32`

GetErrorId returns the ErrorId field if non-nil, zero value otherwise.

### GetErrorIdOk

`func (o *ReceiptError) GetErrorIdOk() (*uint32, bool)`

GetErrorIdOk returns a tuple with the ErrorId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorId

`func (o *ReceiptError) SetErrorId(v uint32)`

SetErrorId sets ErrorId field to given value.


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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


