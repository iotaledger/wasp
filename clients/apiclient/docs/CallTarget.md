# CallTarget

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContractHName** | **string** | The contract name as HName (Hex) | 
**FunctionHName** | **string** | The function name as HName (Hex) | 

## Methods

### NewCallTarget

`func NewCallTarget(contractHName string, functionHName string, ) *CallTarget`

NewCallTarget instantiates a new CallTarget object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCallTargetWithDefaults

`func NewCallTargetWithDefaults() *CallTarget`

NewCallTargetWithDefaults instantiates a new CallTarget object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContractHName

`func (o *CallTarget) GetContractHName() string`

GetContractHName returns the ContractHName field if non-nil, zero value otherwise.

### GetContractHNameOk

`func (o *CallTarget) GetContractHNameOk() (*string, bool)`

GetContractHNameOk returns a tuple with the ContractHName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractHName

`func (o *CallTarget) SetContractHName(v string)`

SetContractHName sets ContractHName field to given value.


### GetFunctionHName

`func (o *CallTarget) GetFunctionHName() string`

GetFunctionHName returns the FunctionHName field if non-nil, zero value otherwise.

### GetFunctionHNameOk

`func (o *CallTarget) GetFunctionHNameOk() (*string, bool)`

GetFunctionHNameOk returns a tuple with the FunctionHName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFunctionHName

`func (o *CallTarget) SetFunctionHName(v string)`

SetFunctionHName sets FunctionHName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


