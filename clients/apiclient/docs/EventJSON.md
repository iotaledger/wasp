# EventJSON

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContractID** | **uint32** | ID of the Contract that issued the event | 
**Payload** | **string** | payload | 
**Timestamp** | **int64** | timestamp | 
**Topic** | **string** | topic | 

## Methods

### NewEventJSON

`func NewEventJSON(contractID uint32, payload string, timestamp int64, topic string, ) *EventJSON`

NewEventJSON instantiates a new EventJSON object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEventJSONWithDefaults

`func NewEventJSONWithDefaults() *EventJSON`

NewEventJSONWithDefaults instantiates a new EventJSON object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContractID

`func (o *EventJSON) GetContractID() uint32`

GetContractID returns the ContractID field if non-nil, zero value otherwise.

### GetContractIDOk

`func (o *EventJSON) GetContractIDOk() (*uint32, bool)`

GetContractIDOk returns a tuple with the ContractID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractID

`func (o *EventJSON) SetContractID(v uint32)`

SetContractID sets ContractID field to given value.


### GetPayload

`func (o *EventJSON) GetPayload() string`

GetPayload returns the Payload field if non-nil, zero value otherwise.

### GetPayloadOk

`func (o *EventJSON) GetPayloadOk() (*string, bool)`

GetPayloadOk returns a tuple with the Payload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPayload

`func (o *EventJSON) SetPayload(v string)`

SetPayload sets Payload field to given value.


### GetTimestamp

`func (o *EventJSON) GetTimestamp() int64`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *EventJSON) GetTimestampOk() (*int64, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *EventJSON) SetTimestamp(v int64)`

SetTimestamp sets Timestamp field to given value.


### GetTopic

`func (o *EventJSON) GetTopic() string`

GetTopic returns the Topic field if non-nil, zero value otherwise.

### GetTopicOk

`func (o *EventJSON) GetTopicOk() (*string, bool)`

GetTopicOk returns a tuple with the Topic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopic

`func (o *EventJSON) SetTopic(v string)`

SetTopic sets Topic field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


