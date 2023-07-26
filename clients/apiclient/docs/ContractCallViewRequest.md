# ContractCallViewRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Arguments** | [**JSONDict**](JSONDict.md) |  | 
**Block** | Pointer to **string** |  | [optional] 
**ContractHName** | **string** | The contract name as HName (Hex) | 
**ContractName** | **string** | The contract name | 
**FunctionHName** | **string** | The function name as HName (Hex) | 
**FunctionName** | **string** | The function name | 

## Methods

### NewContractCallViewRequest

`func NewContractCallViewRequest(arguments JSONDict, contractHName string, contractName string, functionHName string, functionName string, ) *ContractCallViewRequest`

NewContractCallViewRequest instantiates a new ContractCallViewRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewContractCallViewRequestWithDefaults

`func NewContractCallViewRequestWithDefaults() *ContractCallViewRequest`

NewContractCallViewRequestWithDefaults instantiates a new ContractCallViewRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetArguments

`func (o *ContractCallViewRequest) GetArguments() JSONDict`

GetArguments returns the Arguments field if non-nil, zero value otherwise.

### GetArgumentsOk

`func (o *ContractCallViewRequest) GetArgumentsOk() (*JSONDict, bool)`

GetArgumentsOk returns a tuple with the Arguments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArguments

`func (o *ContractCallViewRequest) SetArguments(v JSONDict)`

SetArguments sets Arguments field to given value.


### GetBlock

`func (o *ContractCallViewRequest) GetBlock() string`

GetBlock returns the Block field if non-nil, zero value otherwise.

### GetBlockOk

`func (o *ContractCallViewRequest) GetBlockOk() (*string, bool)`

GetBlockOk returns a tuple with the Block field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlock

`func (o *ContractCallViewRequest) SetBlock(v string)`

SetBlock sets Block field to given value.

### HasBlock

`func (o *ContractCallViewRequest) HasBlock() bool`

HasBlock returns a boolean if a field has been set.

### GetContractHName

`func (o *ContractCallViewRequest) GetContractHName() string`

GetContractHName returns the ContractHName field if non-nil, zero value otherwise.

### GetContractHNameOk

`func (o *ContractCallViewRequest) GetContractHNameOk() (*string, bool)`

GetContractHNameOk returns a tuple with the ContractHName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractHName

`func (o *ContractCallViewRequest) SetContractHName(v string)`

SetContractHName sets ContractHName field to given value.


### GetContractName

`func (o *ContractCallViewRequest) GetContractName() string`

GetContractName returns the ContractName field if non-nil, zero value otherwise.

### GetContractNameOk

`func (o *ContractCallViewRequest) GetContractNameOk() (*string, bool)`

GetContractNameOk returns a tuple with the ContractName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractName

`func (o *ContractCallViewRequest) SetContractName(v string)`

SetContractName sets ContractName field to given value.


### GetFunctionHName

`func (o *ContractCallViewRequest) GetFunctionHName() string`

GetFunctionHName returns the FunctionHName field if non-nil, zero value otherwise.

### GetFunctionHNameOk

`func (o *ContractCallViewRequest) GetFunctionHNameOk() (*string, bool)`

GetFunctionHNameOk returns a tuple with the FunctionHName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFunctionHName

`func (o *ContractCallViewRequest) SetFunctionHName(v string)`

SetFunctionHName sets FunctionHName field to given value.


### GetFunctionName

`func (o *ContractCallViewRequest) GetFunctionName() string`

GetFunctionName returns the FunctionName field if non-nil, zero value otherwise.

### GetFunctionNameOk

`func (o *ContractCallViewRequest) GetFunctionNameOk() (*string, bool)`

GetFunctionNameOk returns a tuple with the FunctionName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFunctionName

`func (o *ContractCallViewRequest) SetFunctionName(v string)`

SetFunctionName sets FunctionName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


