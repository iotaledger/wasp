/*
Wasp API

REST API for the Wasp node

API version: 0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the EventJSON type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &EventJSON{}

// EventJSON struct for EventJSON
type EventJSON struct {
	// ID of the Contract that issued the event
	ContractID uint32 `json:"contractID"`
	// payload
	Payload string `json:"payload"`
	// timestamp
	Timestamp int64 `json:"timestamp"`
	// topic
	Topic string `json:"topic"`
}

type _EventJSON EventJSON

// NewEventJSON instantiates a new EventJSON object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEventJSON(contractID uint32, payload string, timestamp int64, topic string) *EventJSON {
	this := EventJSON{}
	this.ContractID = contractID
	this.Payload = payload
	this.Timestamp = timestamp
	this.Topic = topic
	return &this
}

// NewEventJSONWithDefaults instantiates a new EventJSON object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEventJSONWithDefaults() *EventJSON {
	this := EventJSON{}
	return &this
}

// GetContractID returns the ContractID field value
func (o *EventJSON) GetContractID() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.ContractID
}

// GetContractIDOk returns a tuple with the ContractID field value
// and a boolean to check if the value has been set.
func (o *EventJSON) GetContractIDOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ContractID, true
}

// SetContractID sets field value
func (o *EventJSON) SetContractID(v uint32) {
	o.ContractID = v
}

// GetPayload returns the Payload field value
func (o *EventJSON) GetPayload() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Payload
}

// GetPayloadOk returns a tuple with the Payload field value
// and a boolean to check if the value has been set.
func (o *EventJSON) GetPayloadOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Payload, true
}

// SetPayload sets field value
func (o *EventJSON) SetPayload(v string) {
	o.Payload = v
}

// GetTimestamp returns the Timestamp field value
func (o *EventJSON) GetTimestamp() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.Timestamp
}

// GetTimestampOk returns a tuple with the Timestamp field value
// and a boolean to check if the value has been set.
func (o *EventJSON) GetTimestampOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Timestamp, true
}

// SetTimestamp sets field value
func (o *EventJSON) SetTimestamp(v int64) {
	o.Timestamp = v
}

// GetTopic returns the Topic field value
func (o *EventJSON) GetTopic() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Topic
}

// GetTopicOk returns a tuple with the Topic field value
// and a boolean to check if the value has been set.
func (o *EventJSON) GetTopicOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Topic, true
}

// SetTopic sets field value
func (o *EventJSON) SetTopic(v string) {
	o.Topic = v
}

func (o EventJSON) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o EventJSON) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["contractID"] = o.ContractID
	toSerialize["payload"] = o.Payload
	toSerialize["timestamp"] = o.Timestamp
	toSerialize["topic"] = o.Topic
	return toSerialize, nil
}

func (o *EventJSON) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"contractID",
		"payload",
		"timestamp",
		"topic",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varEventJSON := _EventJSON{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varEventJSON)

	if err != nil {
		return err
	}

	*o = EventJSON(varEventJSON)

	return err
}

type NullableEventJSON struct {
	value *EventJSON
	isSet bool
}

func (v NullableEventJSON) Get() *EventJSON {
	return v.value
}

func (v *NullableEventJSON) Set(val *EventJSON) {
	v.value = val
	v.isSet = true
}

func (v NullableEventJSON) IsSet() bool {
	return v.isSet
}

func (v *NullableEventJSON) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEventJSON(val *EventJSON) *NullableEventJSON {
	return &NullableEventJSON{value: val, isSet: true}
}

func (v NullableEventJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEventJSON) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


