# StateTransaction

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StateIndex** | **uint32** | The state index | 
**TxId** | **string** | The transaction ID | 

## Methods

### NewStateTransaction

`func NewStateTransaction(stateIndex uint32, txId string, ) *StateTransaction`

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

`func (o *StateTransaction) GetStateIndex() uint32`

GetStateIndex returns the StateIndex field if non-nil, zero value otherwise.

### GetStateIndexOk

`func (o *StateTransaction) GetStateIndexOk() (*uint32, bool)`

GetStateIndexOk returns a tuple with the StateIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateIndex

`func (o *StateTransaction) SetStateIndex(v uint32)`

SetStateIndex sets StateIndex field to given value.


### GetTxId

`func (o *StateTransaction) GetTxId() string`

GetTxId returns the TxId field if non-nil, zero value otherwise.

### GetTxIdOk

`func (o *StateTransaction) GetTxIdOk() (*string, bool)`

GetTxIdOk returns a tuple with the TxId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTxId

`func (o *StateTransaction) SetTxId(v string)`

SetTxId sets TxId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


