# RequestDetail

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Allowance** | Pointer to [**Allowance**](Allowance.md) |  | [optional] 
**CallTarget** | Pointer to [**CallTarget**](CallTarget.md) |  | [optional] 
**FungibleTokens** | Pointer to [**FungibleTokens**](FungibleTokens.md) |  | [optional] 
**GasGudget** | Pointer to **int64** |  | [optional] 
**IsEVM** | Pointer to **bool** |  | [optional] 
**IsOffLedger** | Pointer to **bool** |  | [optional] 
**Nft** | Pointer to [**NFTDataResponse**](NFTDataResponse.md) |  | [optional] 
**Params** | Pointer to [**JSONDict**](JSONDict.md) |  | [optional] 
**RequestId** | Pointer to **string** |  | [optional] 
**SenderAccount** | Pointer to **string** |  | [optional] 
**TargetAddress** | Pointer to **string** |  | [optional] 

## Methods

### NewRequestDetail

`func NewRequestDetail() *RequestDetail`

NewRequestDetail instantiates a new RequestDetail object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRequestDetailWithDefaults

`func NewRequestDetailWithDefaults() *RequestDetail`

NewRequestDetailWithDefaults instantiates a new RequestDetail object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowance

`func (o *RequestDetail) GetAllowance() Allowance`

GetAllowance returns the Allowance field if non-nil, zero value otherwise.

### GetAllowanceOk

`func (o *RequestDetail) GetAllowanceOk() (*Allowance, bool)`

GetAllowanceOk returns a tuple with the Allowance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowance

`func (o *RequestDetail) SetAllowance(v Allowance)`

SetAllowance sets Allowance field to given value.

### HasAllowance

`func (o *RequestDetail) HasAllowance() bool`

HasAllowance returns a boolean if a field has been set.

### GetCallTarget

`func (o *RequestDetail) GetCallTarget() CallTarget`

GetCallTarget returns the CallTarget field if non-nil, zero value otherwise.

### GetCallTargetOk

`func (o *RequestDetail) GetCallTargetOk() (*CallTarget, bool)`

GetCallTargetOk returns a tuple with the CallTarget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCallTarget

`func (o *RequestDetail) SetCallTarget(v CallTarget)`

SetCallTarget sets CallTarget field to given value.

### HasCallTarget

`func (o *RequestDetail) HasCallTarget() bool`

HasCallTarget returns a boolean if a field has been set.

### GetFungibleTokens

`func (o *RequestDetail) GetFungibleTokens() FungibleTokens`

GetFungibleTokens returns the FungibleTokens field if non-nil, zero value otherwise.

### GetFungibleTokensOk

`func (o *RequestDetail) GetFungibleTokensOk() (*FungibleTokens, bool)`

GetFungibleTokensOk returns a tuple with the FungibleTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFungibleTokens

`func (o *RequestDetail) SetFungibleTokens(v FungibleTokens)`

SetFungibleTokens sets FungibleTokens field to given value.

### HasFungibleTokens

`func (o *RequestDetail) HasFungibleTokens() bool`

HasFungibleTokens returns a boolean if a field has been set.

### GetGasGudget

`func (o *RequestDetail) GetGasGudget() int64`

GetGasGudget returns the GasGudget field if non-nil, zero value otherwise.

### GetGasGudgetOk

`func (o *RequestDetail) GetGasGudgetOk() (*int64, bool)`

GetGasGudgetOk returns a tuple with the GasGudget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasGudget

`func (o *RequestDetail) SetGasGudget(v int64)`

SetGasGudget sets GasGudget field to given value.

### HasGasGudget

`func (o *RequestDetail) HasGasGudget() bool`

HasGasGudget returns a boolean if a field has been set.

### GetIsEVM

`func (o *RequestDetail) GetIsEVM() bool`

GetIsEVM returns the IsEVM field if non-nil, zero value otherwise.

### GetIsEVMOk

`func (o *RequestDetail) GetIsEVMOk() (*bool, bool)`

GetIsEVMOk returns a tuple with the IsEVM field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsEVM

`func (o *RequestDetail) SetIsEVM(v bool)`

SetIsEVM sets IsEVM field to given value.

### HasIsEVM

`func (o *RequestDetail) HasIsEVM() bool`

HasIsEVM returns a boolean if a field has been set.

### GetIsOffLedger

`func (o *RequestDetail) GetIsOffLedger() bool`

GetIsOffLedger returns the IsOffLedger field if non-nil, zero value otherwise.

### GetIsOffLedgerOk

`func (o *RequestDetail) GetIsOffLedgerOk() (*bool, bool)`

GetIsOffLedgerOk returns a tuple with the IsOffLedger field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsOffLedger

`func (o *RequestDetail) SetIsOffLedger(v bool)`

SetIsOffLedger sets IsOffLedger field to given value.

### HasIsOffLedger

`func (o *RequestDetail) HasIsOffLedger() bool`

HasIsOffLedger returns a boolean if a field has been set.

### GetNft

`func (o *RequestDetail) GetNft() NFTDataResponse`

GetNft returns the Nft field if non-nil, zero value otherwise.

### GetNftOk

`func (o *RequestDetail) GetNftOk() (*NFTDataResponse, bool)`

GetNftOk returns a tuple with the Nft field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNft

`func (o *RequestDetail) SetNft(v NFTDataResponse)`

SetNft sets Nft field to given value.

### HasNft

`func (o *RequestDetail) HasNft() bool`

HasNft returns a boolean if a field has been set.

### GetParams

`func (o *RequestDetail) GetParams() JSONDict`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *RequestDetail) GetParamsOk() (*JSONDict, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *RequestDetail) SetParams(v JSONDict)`

SetParams sets Params field to given value.

### HasParams

`func (o *RequestDetail) HasParams() bool`

HasParams returns a boolean if a field has been set.

### GetRequestId

`func (o *RequestDetail) GetRequestId() string`

GetRequestId returns the RequestId field if non-nil, zero value otherwise.

### GetRequestIdOk

`func (o *RequestDetail) GetRequestIdOk() (*string, bool)`

GetRequestIdOk returns a tuple with the RequestId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestId

`func (o *RequestDetail) SetRequestId(v string)`

SetRequestId sets RequestId field to given value.

### HasRequestId

`func (o *RequestDetail) HasRequestId() bool`

HasRequestId returns a boolean if a field has been set.

### GetSenderAccount

`func (o *RequestDetail) GetSenderAccount() string`

GetSenderAccount returns the SenderAccount field if non-nil, zero value otherwise.

### GetSenderAccountOk

`func (o *RequestDetail) GetSenderAccountOk() (*string, bool)`

GetSenderAccountOk returns a tuple with the SenderAccount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSenderAccount

`func (o *RequestDetail) SetSenderAccount(v string)`

SetSenderAccount sets SenderAccount field to given value.

### HasSenderAccount

`func (o *RequestDetail) HasSenderAccount() bool`

HasSenderAccount returns a boolean if a field has been set.

### GetTargetAddress

`func (o *RequestDetail) GetTargetAddress() string`

GetTargetAddress returns the TargetAddress field if non-nil, zero value otherwise.

### GetTargetAddressOk

`func (o *RequestDetail) GetTargetAddressOk() (*string, bool)`

GetTargetAddressOk returns a tuple with the TargetAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetAddress

`func (o *RequestDetail) SetTargetAddress(v string)`

SetTargetAddress sets TargetAddress field to given value.

### HasTargetAddress

`func (o *RequestDetail) HasTargetAddress() bool`

HasTargetAddress returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


