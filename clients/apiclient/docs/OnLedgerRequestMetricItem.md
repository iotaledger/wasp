# OnLedgerRequestMetricItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastMessage** | [**OnLedgerRequest**](OnLedgerRequest.md) |  | 
**Messages** | **uint32** |  | 
**Timestamp** | **time.Time** |  | 

## Methods

### NewOnLedgerRequestMetricItem

`func NewOnLedgerRequestMetricItem(lastMessage OnLedgerRequest, messages uint32, timestamp time.Time, ) *OnLedgerRequestMetricItem`

NewOnLedgerRequestMetricItem instantiates a new OnLedgerRequestMetricItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOnLedgerRequestMetricItemWithDefaults

`func NewOnLedgerRequestMetricItemWithDefaults() *OnLedgerRequestMetricItem`

NewOnLedgerRequestMetricItemWithDefaults instantiates a new OnLedgerRequestMetricItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastMessage

`func (o *OnLedgerRequestMetricItem) GetLastMessage() OnLedgerRequest`

GetLastMessage returns the LastMessage field if non-nil, zero value otherwise.

### GetLastMessageOk

`func (o *OnLedgerRequestMetricItem) GetLastMessageOk() (*OnLedgerRequest, bool)`

GetLastMessageOk returns a tuple with the LastMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastMessage

`func (o *OnLedgerRequestMetricItem) SetLastMessage(v OnLedgerRequest)`

SetLastMessage sets LastMessage field to given value.


### GetMessages

`func (o *OnLedgerRequestMetricItem) GetMessages() uint32`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *OnLedgerRequestMetricItem) GetMessagesOk() (*uint32, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *OnLedgerRequestMetricItem) SetMessages(v uint32)`

SetMessages sets Messages field to given value.


### GetTimestamp

`func (o *OnLedgerRequestMetricItem) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *OnLedgerRequestMetricItem) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *OnLedgerRequestMetricItem) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


