# Event

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContractID** | Pointer to **int32** |  | [optional] 
**Payload** | Pointer to **[]int32** |  | [optional] 
**Timestamp** | Pointer to **int64** |  | [optional] 
**Topic** | Pointer to **string** |  | [optional] 

## Methods

### NewEvent

`func NewEvent() *Event`

NewEvent instantiates a new Event object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEventWithDefaults

`func NewEventWithDefaults() *Event`

NewEventWithDefaults instantiates a new Event object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContractID

`func (o *Event) GetContractID() int32`

GetContractID returns the ContractID field if non-nil, zero value otherwise.

### GetContractIDOk

`func (o *Event) GetContractIDOk() (*int32, bool)`

GetContractIDOk returns a tuple with the ContractID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractID

`func (o *Event) SetContractID(v int32)`

SetContractID sets ContractID field to given value.

### HasContractID

`func (o *Event) HasContractID() bool`

HasContractID returns a boolean if a field has been set.

### GetPayload

`func (o *Event) GetPayload() []int32`

GetPayload returns the Payload field if non-nil, zero value otherwise.

### GetPayloadOk

`func (o *Event) GetPayloadOk() (*[]int32, bool)`

GetPayloadOk returns a tuple with the Payload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPayload

`func (o *Event) SetPayload(v []int32)`

SetPayload sets Payload field to given value.

### HasPayload

`func (o *Event) HasPayload() bool`

HasPayload returns a boolean if a field has been set.

### GetTimestamp

`func (o *Event) GetTimestamp() int64`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *Event) GetTimestampOk() (*int64, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *Event) SetTimestamp(v int64)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *Event) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.

### GetTopic

`func (o *Event) GetTopic() string`

GetTopic returns the Topic field if non-nil, zero value otherwise.

### GetTopicOk

`func (o *Event) GetTopicOk() (*string, bool)`

GetTopicOk returns a tuple with the Topic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopic

`func (o *Event) SetTopic(v string)`

SetTopic sets Topic field to given value.

### HasTopic

`func (o *Event) HasTopic() bool`

HasTopic returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


