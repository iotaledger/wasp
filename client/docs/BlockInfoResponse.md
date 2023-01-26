# BlockInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AnchorTransactionId** | Pointer to **string** |  | [optional] 
**BlockIndex** | Pointer to **int32** |  | [optional] 
**GasBurned** | Pointer to **int64** |  | [optional] 
**GasFeeCharged** | Pointer to **int64** |  | [optional] 
**L1CommitmentHash** | Pointer to **string** |  | [optional] 
**NumOffLedgerRequests** | Pointer to **int32** |  | [optional] 
**NumSuccessfulRequests** | Pointer to **int32** |  | [optional] 
**PreviousL1CommitmentHash** | Pointer to **string** |  | [optional] 
**Timestamp** | Pointer to **time.Time** |  | [optional] 
**TotalBaseTokensInL2Accounts** | Pointer to **int64** |  | [optional] 
**TotalRequests** | Pointer to **int32** |  | [optional] 
**TotalStorageDeposit** | Pointer to **int64** |  | [optional] 
**TransactionSubEssenceHash** | Pointer to **string** |  | [optional] 

## Methods

### NewBlockInfoResponse

`func NewBlockInfoResponse() *BlockInfoResponse`

NewBlockInfoResponse instantiates a new BlockInfoResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBlockInfoResponseWithDefaults

`func NewBlockInfoResponseWithDefaults() *BlockInfoResponse`

NewBlockInfoResponseWithDefaults instantiates a new BlockInfoResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnchorTransactionId

`func (o *BlockInfoResponse) GetAnchorTransactionId() string`

GetAnchorTransactionId returns the AnchorTransactionId field if non-nil, zero value otherwise.

### GetAnchorTransactionIdOk

`func (o *BlockInfoResponse) GetAnchorTransactionIdOk() (*string, bool)`

GetAnchorTransactionIdOk returns a tuple with the AnchorTransactionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnchorTransactionId

`func (o *BlockInfoResponse) SetAnchorTransactionId(v string)`

SetAnchorTransactionId sets AnchorTransactionId field to given value.

### HasAnchorTransactionId

`func (o *BlockInfoResponse) HasAnchorTransactionId() bool`

HasAnchorTransactionId returns a boolean if a field has been set.

### GetBlockIndex

`func (o *BlockInfoResponse) GetBlockIndex() int32`

GetBlockIndex returns the BlockIndex field if non-nil, zero value otherwise.

### GetBlockIndexOk

`func (o *BlockInfoResponse) GetBlockIndexOk() (*int32, bool)`

GetBlockIndexOk returns a tuple with the BlockIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlockIndex

`func (o *BlockInfoResponse) SetBlockIndex(v int32)`

SetBlockIndex sets BlockIndex field to given value.

### HasBlockIndex

`func (o *BlockInfoResponse) HasBlockIndex() bool`

HasBlockIndex returns a boolean if a field has been set.

### GetGasBurned

`func (o *BlockInfoResponse) GetGasBurned() int64`

GetGasBurned returns the GasBurned field if non-nil, zero value otherwise.

### GetGasBurnedOk

`func (o *BlockInfoResponse) GetGasBurnedOk() (*int64, bool)`

GetGasBurnedOk returns a tuple with the GasBurned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBurned

`func (o *BlockInfoResponse) SetGasBurned(v int64)`

SetGasBurned sets GasBurned field to given value.

### HasGasBurned

`func (o *BlockInfoResponse) HasGasBurned() bool`

HasGasBurned returns a boolean if a field has been set.

### GetGasFeeCharged

`func (o *BlockInfoResponse) GetGasFeeCharged() int64`

GetGasFeeCharged returns the GasFeeCharged field if non-nil, zero value otherwise.

### GetGasFeeChargedOk

`func (o *BlockInfoResponse) GetGasFeeChargedOk() (*int64, bool)`

GetGasFeeChargedOk returns a tuple with the GasFeeCharged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeeCharged

`func (o *BlockInfoResponse) SetGasFeeCharged(v int64)`

SetGasFeeCharged sets GasFeeCharged field to given value.

### HasGasFeeCharged

`func (o *BlockInfoResponse) HasGasFeeCharged() bool`

HasGasFeeCharged returns a boolean if a field has been set.

### GetL1CommitmentHash

`func (o *BlockInfoResponse) GetL1CommitmentHash() string`

GetL1CommitmentHash returns the L1CommitmentHash field if non-nil, zero value otherwise.

### GetL1CommitmentHashOk

`func (o *BlockInfoResponse) GetL1CommitmentHashOk() (*string, bool)`

GetL1CommitmentHashOk returns a tuple with the L1CommitmentHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetL1CommitmentHash

`func (o *BlockInfoResponse) SetL1CommitmentHash(v string)`

SetL1CommitmentHash sets L1CommitmentHash field to given value.

### HasL1CommitmentHash

`func (o *BlockInfoResponse) HasL1CommitmentHash() bool`

HasL1CommitmentHash returns a boolean if a field has been set.

### GetNumOffLedgerRequests

`func (o *BlockInfoResponse) GetNumOffLedgerRequests() int32`

GetNumOffLedgerRequests returns the NumOffLedgerRequests field if non-nil, zero value otherwise.

### GetNumOffLedgerRequestsOk

`func (o *BlockInfoResponse) GetNumOffLedgerRequestsOk() (*int32, bool)`

GetNumOffLedgerRequestsOk returns a tuple with the NumOffLedgerRequests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumOffLedgerRequests

`func (o *BlockInfoResponse) SetNumOffLedgerRequests(v int32)`

SetNumOffLedgerRequests sets NumOffLedgerRequests field to given value.

### HasNumOffLedgerRequests

`func (o *BlockInfoResponse) HasNumOffLedgerRequests() bool`

HasNumOffLedgerRequests returns a boolean if a field has been set.

### GetNumSuccessfulRequests

`func (o *BlockInfoResponse) GetNumSuccessfulRequests() int32`

GetNumSuccessfulRequests returns the NumSuccessfulRequests field if non-nil, zero value otherwise.

### GetNumSuccessfulRequestsOk

`func (o *BlockInfoResponse) GetNumSuccessfulRequestsOk() (*int32, bool)`

GetNumSuccessfulRequestsOk returns a tuple with the NumSuccessfulRequests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumSuccessfulRequests

`func (o *BlockInfoResponse) SetNumSuccessfulRequests(v int32)`

SetNumSuccessfulRequests sets NumSuccessfulRequests field to given value.

### HasNumSuccessfulRequests

`func (o *BlockInfoResponse) HasNumSuccessfulRequests() bool`

HasNumSuccessfulRequests returns a boolean if a field has been set.

### GetPreviousL1CommitmentHash

`func (o *BlockInfoResponse) GetPreviousL1CommitmentHash() string`

GetPreviousL1CommitmentHash returns the PreviousL1CommitmentHash field if non-nil, zero value otherwise.

### GetPreviousL1CommitmentHashOk

`func (o *BlockInfoResponse) GetPreviousL1CommitmentHashOk() (*string, bool)`

GetPreviousL1CommitmentHashOk returns a tuple with the PreviousL1CommitmentHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreviousL1CommitmentHash

`func (o *BlockInfoResponse) SetPreviousL1CommitmentHash(v string)`

SetPreviousL1CommitmentHash sets PreviousL1CommitmentHash field to given value.

### HasPreviousL1CommitmentHash

`func (o *BlockInfoResponse) HasPreviousL1CommitmentHash() bool`

HasPreviousL1CommitmentHash returns a boolean if a field has been set.

### GetTimestamp

`func (o *BlockInfoResponse) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *BlockInfoResponse) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *BlockInfoResponse) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *BlockInfoResponse) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.

### GetTotalBaseTokensInL2Accounts

`func (o *BlockInfoResponse) GetTotalBaseTokensInL2Accounts() int64`

GetTotalBaseTokensInL2Accounts returns the TotalBaseTokensInL2Accounts field if non-nil, zero value otherwise.

### GetTotalBaseTokensInL2AccountsOk

`func (o *BlockInfoResponse) GetTotalBaseTokensInL2AccountsOk() (*int64, bool)`

GetTotalBaseTokensInL2AccountsOk returns a tuple with the TotalBaseTokensInL2Accounts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalBaseTokensInL2Accounts

`func (o *BlockInfoResponse) SetTotalBaseTokensInL2Accounts(v int64)`

SetTotalBaseTokensInL2Accounts sets TotalBaseTokensInL2Accounts field to given value.

### HasTotalBaseTokensInL2Accounts

`func (o *BlockInfoResponse) HasTotalBaseTokensInL2Accounts() bool`

HasTotalBaseTokensInL2Accounts returns a boolean if a field has been set.

### GetTotalRequests

`func (o *BlockInfoResponse) GetTotalRequests() int32`

GetTotalRequests returns the TotalRequests field if non-nil, zero value otherwise.

### GetTotalRequestsOk

`func (o *BlockInfoResponse) GetTotalRequestsOk() (*int32, bool)`

GetTotalRequestsOk returns a tuple with the TotalRequests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequests

`func (o *BlockInfoResponse) SetTotalRequests(v int32)`

SetTotalRequests sets TotalRequests field to given value.

### HasTotalRequests

`func (o *BlockInfoResponse) HasTotalRequests() bool`

HasTotalRequests returns a boolean if a field has been set.

### GetTotalStorageDeposit

`func (o *BlockInfoResponse) GetTotalStorageDeposit() int64`

GetTotalStorageDeposit returns the TotalStorageDeposit field if non-nil, zero value otherwise.

### GetTotalStorageDepositOk

`func (o *BlockInfoResponse) GetTotalStorageDepositOk() (*int64, bool)`

GetTotalStorageDepositOk returns a tuple with the TotalStorageDeposit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalStorageDeposit

`func (o *BlockInfoResponse) SetTotalStorageDeposit(v int64)`

SetTotalStorageDeposit sets TotalStorageDeposit field to given value.

### HasTotalStorageDeposit

`func (o *BlockInfoResponse) HasTotalStorageDeposit() bool`

HasTotalStorageDeposit returns a boolean if a field has been set.

### GetTransactionSubEssenceHash

`func (o *BlockInfoResponse) GetTransactionSubEssenceHash() string`

GetTransactionSubEssenceHash returns the TransactionSubEssenceHash field if non-nil, zero value otherwise.

### GetTransactionSubEssenceHashOk

`func (o *BlockInfoResponse) GetTransactionSubEssenceHashOk() (*string, bool)`

GetTransactionSubEssenceHashOk returns a tuple with the TransactionSubEssenceHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionSubEssenceHash

`func (o *BlockInfoResponse) SetTransactionSubEssenceHash(v string)`

SetTransactionSubEssenceHash sets TransactionSubEssenceHash field to given value.

### HasTransactionSubEssenceHash

`func (o *BlockInfoResponse) HasTransactionSubEssenceHash() bool`

HasTransactionSubEssenceHash returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


