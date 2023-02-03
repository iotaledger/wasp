# RequestDetail

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Allowance** | [**Assets**](Assets.md) |  | 
**CallTarget** | [**CallTarget**](CallTarget.md) |  | 
**FungibleTokens** | [**Assets**](Assets.md) |  | 
**GasGudget** | **int64** |  | 
**IsEVM** | **bool** |  | 
**IsOffLedger** | **bool** |  | 
**Nft** | [**NFTDataResponse**](NFTDataResponse.md) |  | 
**Params** | [**JSONDict**](JSONDict.md) |  | 
**RequestId** | **string** |  | 
**SenderAccount** | **string** |  | 
**TargetAddress** | **string** |  | 

## Methods

### NewRequestDetail

`func NewRequestDetail(allowance Assets, callTarget CallTarget, fungibleTokens Assets, gasGudget int64, isEVM bool, isOffLedger bool, nft NFTDataResponse, params JSONDict, requestId string, senderAccount string, targetAddress string, ) *RequestDetail`

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

`func (o *RequestDetail) GetAllowance() Assets`

GetAllowance returns the Allowance field if non-nil, zero value otherwise.

### GetAllowanceOk

`func (o *RequestDetail) GetAllowanceOk() (*Assets, bool)`

GetAllowanceOk returns a tuple with the Allowance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowance

`func (o *RequestDetail) SetAllowance(v Assets)`

SetAllowance sets Allowance field to given value.


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


### GetFungibleTokens

`func (o *RequestDetail) GetFungibleTokens() Assets`

GetFungibleTokens returns the FungibleTokens field if non-nil, zero value otherwise.

### GetFungibleTokensOk

`func (o *RequestDetail) GetFungibleTokensOk() (*Assets, bool)`

GetFungibleTokensOk returns a tuple with the FungibleTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFungibleTokens

`func (o *RequestDetail) SetFungibleTokens(v Assets)`

SetFungibleTokens sets FungibleTokens field to given value.


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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


