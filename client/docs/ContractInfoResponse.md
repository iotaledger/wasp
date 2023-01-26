# ContractInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** | The description of the contract. | [optional] 
**HName** | Pointer to **string** | The id (HName as Hex)) of the contract. | [optional] 
**Name** | Pointer to **string** | The name of the contract. | [optional] 
**ProgramHash** | Pointer to **[]int32** | The hash of the contract. | [optional] 

## Methods

### NewContractInfoResponse

`func NewContractInfoResponse() *ContractInfoResponse`

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

### HasDescription

`func (o *ContractInfoResponse) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

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

### HasHName

`func (o *ContractInfoResponse) HasHName() bool`

HasHName returns a boolean if a field has been set.

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

### HasName

`func (o *ContractInfoResponse) HasName() bool`

HasName returns a boolean if a field has been set.

### GetProgramHash

`func (o *ContractInfoResponse) GetProgramHash() []int32`

GetProgramHash returns the ProgramHash field if non-nil, zero value otherwise.

### GetProgramHashOk

`func (o *ContractInfoResponse) GetProgramHashOk() (*[]int32, bool)`

GetProgramHashOk returns a tuple with the ProgramHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProgramHash

`func (o *ContractInfoResponse) SetProgramHash(v []int32)`

SetProgramHash sets ProgramHash field to given value.

### HasProgramHash

`func (o *ContractInfoResponse) HasProgramHash() bool`

HasProgramHash returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


