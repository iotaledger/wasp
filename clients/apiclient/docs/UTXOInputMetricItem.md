# UTXOInputMetricItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastMessage** | [**OutputID**](OutputID.md) |  | 
**Messages** | **uint32** |  | 
**Timestamp** | **time.Time** |  | 

## Methods

### NewUTXOInputMetricItem

`func NewUTXOInputMetricItem(lastMessage OutputID, messages uint32, timestamp time.Time, ) *UTXOInputMetricItem`

NewUTXOInputMetricItem instantiates a new UTXOInputMetricItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUTXOInputMetricItemWithDefaults

`func NewUTXOInputMetricItemWithDefaults() *UTXOInputMetricItem`

NewUTXOInputMetricItemWithDefaults instantiates a new UTXOInputMetricItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastMessage

`func (o *UTXOInputMetricItem) GetLastMessage() OutputID`

GetLastMessage returns the LastMessage field if non-nil, zero value otherwise.

### GetLastMessageOk

`func (o *UTXOInputMetricItem) GetLastMessageOk() (*OutputID, bool)`

GetLastMessageOk returns a tuple with the LastMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastMessage

`func (o *UTXOInputMetricItem) SetLastMessage(v OutputID)`

SetLastMessage sets LastMessage field to given value.


### GetMessages

`func (o *UTXOInputMetricItem) GetMessages() uint32`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *UTXOInputMetricItem) GetMessagesOk() (*uint32, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *UTXOInputMetricItem) SetMessages(v uint32)`

SetMessages sets Messages field to given value.


### GetTimestamp

`func (o *UTXOInputMetricItem) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *UTXOInputMetricItem) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *UTXOInputMetricItem) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


