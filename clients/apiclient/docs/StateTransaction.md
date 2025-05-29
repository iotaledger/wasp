# StateTransaction

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StateIndex** | **int32** | The state index | 
**TxDigest** | **string** | The transaction Digest | 

## Methods

### NewStateTransaction

`func NewStateTransaction(stateIndex int32, txDigest string, ) *StateTransaction`

NewStateTransaction instantiates a new StateTransaction object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStateTransactionWithDefaults

`func NewStateTransactionWithDefaults() *StateTransaction`

NewStateTransactionWithDefaults instantiates a new StateTransaction object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStateIndex

`func (o *StateTransaction) GetStateIndex() int32`

GetStateIndex returns the StateIndex field if non-nil, zero value otherwise.

### GetStateIndexOk

`func (o *StateTransaction) GetStateIndexOk() (*int32, bool)`

GetStateIndexOk returns a tuple with the StateIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateIndex

`func (o *StateTransaction) SetStateIndex(v int32)`

SetStateIndex sets StateIndex field to given value.


### GetTxDigest

`func (o *StateTransaction) GetTxDigest() string`

GetTxDigest returns the TxDigest field if non-nil, zero value otherwise.

### GetTxDigestOk

`func (o *StateTransaction) GetTxDigestOk() (*string, bool)`

GetTxDigestOk returns a tuple with the TxDigest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTxDigest

`func (o *StateTransaction) SetTxDigest(v string)`

SetTxDigest sets TxDigest field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


