# ContractInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | **string** | The description of the contract. | 
**HName** | **string** | The id (HName as Hex)) of the contract. | 
**Name** | **string** | The name of the contract. | 
**ProgramHash** | **string** | The hash of the contract. (Hex encoded) | 

## Methods

### NewContractInfoResponse

`func NewContractInfoResponse(description string, hName string, name string, programHash string, ) *ContractInfoResponse`

NewContractInfoResponse instantiates a new ContractInfoResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewContractInfoResponseWithDefaults

`func NewContractInfoResponseWithDefaults() *ContractInfoResponse`

NewContractInfoResponseWithDefaults instantiates a new ContractInfoResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *ContractInfoResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ContractInfoResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ContractInfoResponse) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetHName

`func (o *ContractInfoResponse) GetHName() string`

GetHName returns the HName field if non-nil, zero value otherwise.

### GetHNameOk

`func (o *ContractInfoResponse) GetHNameOk() (*string, bool)`

GetHNameOk returns a tuple with the HName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHName

`func (o *ContractInfoResponse) SetHName(v string)`

SetHName sets HName field to given value.


### GetName

`func (o *ContractInfoResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ContractInfoResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ContractInfoResponse) SetName(v string)`

SetName sets Name field to given value.


### GetProgramHash

`func (o *ContractInfoResponse) GetProgramHash() string`

GetProgramHash returns the ProgramHash field if non-nil, zero value otherwise.

### GetProgramHashOk

`func (o *ContractInfoResponse) GetProgramHashOk() (*string, bool)`

GetProgramHashOk returns a tuple with the ProgramHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProgramHash

`func (o *ContractInfoResponse) SetProgramHash(v string)`

SetProgramHash sets ProgramHash field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


