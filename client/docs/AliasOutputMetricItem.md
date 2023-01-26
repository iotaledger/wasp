# AliasOutputMetricItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastMessage** | Pointer to [**Output**](Output.md) |  | [optional] 
**Messages** | Pointer to **int32** |  | [optional] 
**Timestamp** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewAliasOutputMetricItem

`func NewAliasOutputMetricItem() *AliasOutputMetricItem`

NewAliasOutputMetricItem instantiates a new AliasOutputMetricItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAliasOutputMetricItemWithDefaults

`func NewAliasOutputMetricItemWithDefaults() *AliasOutputMetricItem`

NewAliasOutputMetricItemWithDefaults instantiates a new AliasOutputMetricItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastMessage

`func (o *AliasOutputMetricItem) GetLastMessage() Output`

GetLastMessage returns the LastMessage field if non-nil, zero value otherwise.

### GetLastMessageOk

`func (o *AliasOutputMetricItem) GetLastMessageOk() (*Output, bool)`

GetLastMessageOk returns a tuple with the LastMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastMessage

`func (o *AliasOutputMetricItem) SetLastMessage(v Output)`

SetLastMessage sets LastMessage field to given value.

### HasLastMessage

`func (o *AliasOutputMetricItem) HasLastMessage() bool`

HasLastMessage returns a boolean if a field has been set.

### GetMessages

`func (o *AliasOutputMetricItem) GetMessages() int32`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *AliasOutputMetricItem) GetMessagesOk() (*int32, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *AliasOutputMetricItem) SetMessages(v int32)`

SetMessages sets Messages field to given value.

### HasMessages

`func (o *AliasOutputMetricItem) HasMessages() bool`

HasMessages returns a boolean if a field has been set.

### GetTimestamp

`func (o *AliasOutputMetricItem) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *AliasOutputMetricItem) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *AliasOutputMetricItem) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *AliasOutputMetricItem) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


