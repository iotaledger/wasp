# TxInclusionStateMsgMetricItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastMessage** | [**TxInclusionStateMsg**](TxInclusionStateMsg.md) |  | 
**Messages** | **uint32** |  | 
**Timestamp** | **time.Time** |  | 

## Methods

### NewTxInclusionStateMsgMetricItem

`func NewTxInclusionStateMsgMetricItem(lastMessage TxInclusionStateMsg, messages uint32, timestamp time.Time, ) *TxInclusionStateMsgMetricItem`

NewTxInclusionStateMsgMetricItem instantiates a new TxInclusionStateMsgMetricItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTxInclusionStateMsgMetricItemWithDefaults

`func NewTxInclusionStateMsgMetricItemWithDefaults() *TxInclusionStateMsgMetricItem`

NewTxInclusionStateMsgMetricItemWithDefaults instantiates a new TxInclusionStateMsgMetricItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastMessage

`func (o *TxInclusionStateMsgMetricItem) GetLastMessage() TxInclusionStateMsg`

GetLastMessage returns the LastMessage field if non-nil, zero value otherwise.

### GetLastMessageOk

`func (o *TxInclusionStateMsgMetricItem) GetLastMessageOk() (*TxInclusionStateMsg, bool)`

GetLastMessageOk returns a tuple with the LastMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastMessage

`func (o *TxInclusionStateMsgMetricItem) SetLastMessage(v TxInclusionStateMsg)`

SetLastMessage sets LastMessage field to given value.


### GetMessages

`func (o *TxInclusionStateMsgMetricItem) GetMessages() uint32`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *TxInclusionStateMsgMetricItem) GetMessagesOk() (*uint32, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *TxInclusionStateMsgMetricItem) SetMessages(v uint32)`

SetMessages sets Messages field to given value.


### GetTimestamp

`func (o *TxInclusionStateMsgMetricItem) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *TxInclusionStateMsgMetricItem) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *TxInclusionStateMsgMetricItem) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


