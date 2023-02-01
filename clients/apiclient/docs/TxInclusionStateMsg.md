# TxInclusionStateMsg

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | **string** | The inclusion state | 
**TxId** | **string** | The transaction ID | 

## Methods

### NewTxInclusionStateMsg

`func NewTxInclusionStateMsg(state string, txId string, ) *TxInclusionStateMsg`

NewTxInclusionStateMsg instantiates a new TxInclusionStateMsg object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTxInclusionStateMsgWithDefaults

`func NewTxInclusionStateMsgWithDefaults() *TxInclusionStateMsg`

NewTxInclusionStateMsgWithDefaults instantiates a new TxInclusionStateMsg object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *TxInclusionStateMsg) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *TxInclusionStateMsg) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *TxInclusionStateMsg) SetState(v string)`

SetState sets State field to given value.


### GetTxId

`func (o *TxInclusionStateMsg) GetTxId() string`

GetTxId returns the TxId field if non-nil, zero value otherwise.

### GetTxIdOk

`func (o *TxInclusionStateMsg) GetTxIdOk() (*string, bool)`

GetTxIdOk returns a tuple with the TxId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTxId

`func (o *TxInclusionStateMsg) SetTxId(v string)`

SetTxId sets TxId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


