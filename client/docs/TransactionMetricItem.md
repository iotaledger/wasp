# TransactionMetricItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastMessage** | Pointer to [**Transaction**](Transaction.md) |  | [optional] 
**Messages** | Pointer to **int32** |  | [optional] 
**Timestamp** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewTransactionMetricItem

`func NewTransactionMetricItem() *TransactionMetricItem`

NewTransactionMetricItem instantiates a new TransactionMetricItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTransactionMetricItemWithDefaults

`func NewTransactionMetricItemWithDefaults() *TransactionMetricItem`

NewTransactionMetricItemWithDefaults instantiates a new TransactionMetricItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastMessage

`func (o *TransactionMetricItem) GetLastMessage() Transaction`

GetLastMessage returns the LastMessage field if non-nil, zero value otherwise.

### GetLastMessageOk

`func (o *TransactionMetricItem) GetLastMessageOk() (*Transaction, bool)`

GetLastMessageOk returns a tuple with the LastMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastMessage

`func (o *TransactionMetricItem) SetLastMessage(v Transaction)`

SetLastMessage sets LastMessage field to given value.

### HasLastMessage

`func (o *TransactionMetricItem) HasLastMessage() bool`

HasLastMessage returns a boolean if a field has been set.

### GetMessages

`func (o *TransactionMetricItem) GetMessages() int32`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *TransactionMetricItem) GetMessagesOk() (*int32, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *TransactionMetricItem) SetMessages(v int32)`

SetMessages sets Messages field to given value.

### HasMessages

`func (o *TransactionMetricItem) HasMessages() bool`

HasMessages returns a boolean if a field has been set.

### GetTimestamp

`func (o *TransactionMetricItem) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *TransactionMetricItem) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *TransactionMetricItem) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *TransactionMetricItem) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


