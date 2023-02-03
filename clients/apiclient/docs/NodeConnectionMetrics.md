# NodeConnectionMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InMilestone** | Pointer to [**NodeConnectionMessageMetrics**](NodeConnectionMessageMetrics.md) |  | [optional] 
**NodeConnectionMessagesMetrics** | Pointer to [**NodeConnectionMessagesMetrics**](NodeConnectionMessagesMetrics.md) |  | [optional] 
**Registered** | Pointer to **[]string** | Chain IDs of the chains registered to receiving L1 events | [optional] 

## Methods

### NewNodeConnectionMetrics

`func NewNodeConnectionMetrics() *NodeConnectionMetrics`

NewNodeConnectionMetrics instantiates a new NodeConnectionMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeConnectionMetricsWithDefaults

`func NewNodeConnectionMetricsWithDefaults() *NodeConnectionMetrics`

NewNodeConnectionMetricsWithDefaults instantiates a new NodeConnectionMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInMilestone

`func (o *NodeConnectionMetrics) GetInMilestone() NodeConnectionMessageMetrics`

GetInMilestone returns the InMilestone field if non-nil, zero value otherwise.

### GetInMilestoneOk

`func (o *NodeConnectionMetrics) GetInMilestoneOk() (*NodeConnectionMessageMetrics, bool)`

GetInMilestoneOk returns a tuple with the InMilestone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInMilestone

`func (o *NodeConnectionMetrics) SetInMilestone(v NodeConnectionMessageMetrics)`

SetInMilestone sets InMilestone field to given value.

### HasInMilestone

`func (o *NodeConnectionMetrics) HasInMilestone() bool`

HasInMilestone returns a boolean if a field has been set.

### GetNodeConnectionMessagesMetrics

`func (o *NodeConnectionMetrics) GetNodeConnectionMessagesMetrics() NodeConnectionMessagesMetrics`

GetNodeConnectionMessagesMetrics returns the NodeConnectionMessagesMetrics field if non-nil, zero value otherwise.

### GetNodeConnectionMessagesMetricsOk

`func (o *NodeConnectionMetrics) GetNodeConnectionMessagesMetricsOk() (*NodeConnectionMessagesMetrics, bool)`

GetNodeConnectionMessagesMetricsOk returns a tuple with the NodeConnectionMessagesMetrics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeConnectionMessagesMetrics

`func (o *NodeConnectionMetrics) SetNodeConnectionMessagesMetrics(v NodeConnectionMessagesMetrics)`

SetNodeConnectionMessagesMetrics sets NodeConnectionMessagesMetrics field to given value.

### HasNodeConnectionMessagesMetrics

`func (o *NodeConnectionMetrics) HasNodeConnectionMessagesMetrics() bool`

HasNodeConnectionMessagesMetrics returns a boolean if a field has been set.

### GetRegistered

`func (o *NodeConnectionMetrics) GetRegistered() []string`

GetRegistered returns the Registered field if non-nil, zero value otherwise.

### GetRegisteredOk

`func (o *NodeConnectionMetrics) GetRegisteredOk() (*[]string, bool)`

GetRegisteredOk returns a tuple with the Registered field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegistered

`func (o *NodeConnectionMetrics) SetRegistered(v []string)`

SetRegistered sets Registered field to given value.

### HasRegistered

`func (o *NodeConnectionMetrics) HasRegistered() bool`

HasRegistered returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


