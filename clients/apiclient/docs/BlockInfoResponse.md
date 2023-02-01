# BlockInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AnchorTransactionId** | **string** |  | 
**BlockIndex** | **uint32** |  | 
**GasBurned** | **int64** |  | 
**GasFeeCharged** | **int64** |  | 
**L1CommitmentHash** | **string** |  | 
**NumOffLedgerRequests** | **uint32** |  | 
**NumSuccessfulRequests** | **uint32** |  | 
**PreviousL1CommitmentHash** | **string** |  | 
**Timestamp** | **time.Time** |  | 
**TotalBaseTokensInL2Accounts** | **int64** |  | 
**TotalRequests** | **uint32** |  | 
**TotalStorageDeposit** | **int64** |  | 
**TransactionSubEssenceHash** | **string** |  | 

## Methods

### NewBlockInfoResponse

`func NewBlockInfoResponse(anchorTransactionId string, blockIndex uint32, gasBurned int64, gasFeeCharged int64, l1CommitmentHash string, numOffLedgerRequests uint32, numSuccessfulRequests uint32, previousL1CommitmentHash string, timestamp time.Time, totalBaseTokensInL2Accounts int64, totalRequests uint32, totalStorageDeposit int64, transactionSubEssenceHash string, ) *BlockInfoResponse`

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


### GetBlockIndex

`func (o *BlockInfoResponse) GetBlockIndex() uint32`

GetBlockIndex returns the BlockIndex field if non-nil, zero value otherwise.

### GetBlockIndexOk

`func (o *BlockInfoResponse) GetBlockIndexOk() (*uint32, bool)`

GetBlockIndexOk returns a tuple with the BlockIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlockIndex

`func (o *BlockInfoResponse) SetBlockIndex(v uint32)`

SetBlockIndex sets BlockIndex field to given value.


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


### GetNumOffLedgerRequests

`func (o *BlockInfoResponse) GetNumOffLedgerRequests() uint32`

GetNumOffLedgerRequests returns the NumOffLedgerRequests field if non-nil, zero value otherwise.

### GetNumOffLedgerRequestsOk

`func (o *BlockInfoResponse) GetNumOffLedgerRequestsOk() (*uint32, bool)`

GetNumOffLedgerRequestsOk returns a tuple with the NumOffLedgerRequests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumOffLedgerRequests

`func (o *BlockInfoResponse) SetNumOffLedgerRequests(v uint32)`

SetNumOffLedgerRequests sets NumOffLedgerRequests field to given value.


### GetNumSuccessfulRequests

`func (o *BlockInfoResponse) GetNumSuccessfulRequests() uint32`

GetNumSuccessfulRequests returns the NumSuccessfulRequests field if non-nil, zero value otherwise.

### GetNumSuccessfulRequestsOk

`func (o *BlockInfoResponse) GetNumSuccessfulRequestsOk() (*uint32, bool)`

GetNumSuccessfulRequestsOk returns a tuple with the NumSuccessfulRequests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumSuccessfulRequests

`func (o *BlockInfoResponse) SetNumSuccessfulRequests(v uint32)`

SetNumSuccessfulRequests sets NumSuccessfulRequests field to given value.


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


### GetTotalRequests

`func (o *BlockInfoResponse) GetTotalRequests() uint32`

GetTotalRequests returns the TotalRequests field if non-nil, zero value otherwise.

### GetTotalRequestsOk

`func (o *BlockInfoResponse) GetTotalRequestsOk() (*uint32, bool)`

GetTotalRequestsOk returns a tuple with the TotalRequests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequests

`func (o *BlockInfoResponse) SetTotalRequests(v uint32)`

SetTotalRequests sets TotalRequests field to given value.


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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


