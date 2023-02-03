# NodeConnectionMessageMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastEvent** | Pointer to **time.Time** | Last time the message was sent/received | [optional] 
**LastMessage** | Pointer to **string** | The print out of the last message | [optional] 
**Total** | Pointer to **int32** | Total number of messages sent/received | [optional] 

## Methods

### NewNodeConnectionMessageMetrics

`func NewNodeConnectionMessageMetrics() *NodeConnectionMessageMetrics`

NewNodeConnectionMessageMetrics instantiates a new NodeConnectionMessageMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeConnectionMessageMetricsWithDefaults

`func NewNodeConnectionMessageMetricsWithDefaults() *NodeConnectionMessageMetrics`

NewNodeConnectionMessageMetricsWithDefaults instantiates a new NodeConnectionMessageMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastEvent

`func (o *NodeConnectionMessageMetrics) GetLastEvent() time.Time`

GetLastEvent returns the LastEvent field if non-nil, zero value otherwise.

### GetLastEventOk

`func (o *NodeConnectionMessageMetrics) GetLastEventOk() (*time.Time, bool)`

GetLastEventOk returns a tuple with the LastEvent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastEvent

`func (o *NodeConnectionMessageMetrics) SetLastEvent(v time.Time)`

SetLastEvent sets LastEvent field to given value.

### HasLastEvent

`func (o *NodeConnectionMessageMetrics) HasLastEvent() bool`

HasLastEvent returns a boolean if a field has been set.

### GetLastMessage

`func (o *NodeConnectionMessageMetrics) GetLastMessage() string`

GetLastMessage returns the LastMessage field if non-nil, zero value otherwise.

### GetLastMessageOk

`func (o *NodeConnectionMessageMetrics) GetLastMessageOk() (*string, bool)`

GetLastMessageOk returns a tuple with the LastMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastMessage

`func (o *NodeConnectionMessageMetrics) SetLastMessage(v string)`

SetLastMessage sets LastMessage field to given value.

### HasLastMessage

`func (o *NodeConnectionMessageMetrics) HasLastMessage() bool`

HasLastMessage returns a boolean if a field has been set.

### GetTotal

`func (o *NodeConnectionMessageMetrics) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *NodeConnectionMessageMetrics) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *NodeConnectionMessageMetrics) SetTotal(v int32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *NodeConnectionMessageMetrics) HasTotal() bool`

HasTotal returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


