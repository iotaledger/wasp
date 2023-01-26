# TxInclusionStateMsg

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** | The inclusion state | [optional] 
**TxId** | Pointer to **string** | The transaction ID | [optional] 

## Methods

### NewTxInclusionStateMsg

`func NewTxInclusionStateMsg() *TxInclusionStateMsg`

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

### HasState

`func (o *TxInclusionStateMsg) HasState() bool`

HasState returns a boolean if a field has been set.

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

### HasTxId

`func (o *TxInclusionStateMsg) HasTxId() bool`

HasTxId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


