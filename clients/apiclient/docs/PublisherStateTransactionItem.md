# PublisherStateTransactionItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastMessage** | [**StateTransaction**](StateTransaction.md) |  | 
**Messages** | **uint32** |  | 
**Timestamp** | **time.Time** |  | 

## Methods

### NewPublisherStateTransactionItem

`func NewPublisherStateTransactionItem(lastMessage StateTransaction, messages uint32, timestamp time.Time, ) *PublisherStateTransactionItem`

NewPublisherStateTransactionItem instantiates a new PublisherStateTransactionItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPublisherStateTransactionItemWithDefaults

`func NewPublisherStateTransactionItemWithDefaults() *PublisherStateTransactionItem`

NewPublisherStateTransactionItemWithDefaults instantiates a new PublisherStateTransactionItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastMessage

`func (o *PublisherStateTransactionItem) GetLastMessage() StateTransaction`

GetLastMessage returns the LastMessage field if non-nil, zero value otherwise.

### GetLastMessageOk

`func (o *PublisherStateTransactionItem) GetLastMessageOk() (*StateTransaction, bool)`

GetLastMessageOk returns a tuple with the LastMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastMessage

`func (o *PublisherStateTransactionItem) SetLastMessage(v StateTransaction)`

SetLastMessage sets LastMessage field to given value.


### GetMessages

`func (o *PublisherStateTransactionItem) GetMessages() uint32`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *PublisherStateTransactionItem) GetMessagesOk() (*uint32, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *PublisherStateTransactionItem) SetMessages(v uint32)`

SetMessages sets Messages field to given value.


### GetTimestamp

`func (o *PublisherStateTransactionItem) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *PublisherStateTransactionItem) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *PublisherStateTransactionItem) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


