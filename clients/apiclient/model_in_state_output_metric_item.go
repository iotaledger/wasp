/*
Wasp API

REST API for the Wasp node

API version: 0.4.0-alpha.8-16-g83edf92b9
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
	"time"
)

// checks if the InStateOutputMetricItem type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &InStateOutputMetricItem{}

// InStateOutputMetricItem struct for InStateOutputMetricItem
type InStateOutputMetricItem struct {
	LastMessage InStateOutput `json:"lastMessage"`
	Messages uint32 `json:"messages"`
	Timestamp time.Time `json:"timestamp"`
}

// NewInStateOutputMetricItem instantiates a new InStateOutputMetricItem object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewInStateOutputMetricItem(lastMessage InStateOutput, messages uint32, timestamp time.Time) *InStateOutputMetricItem {
	this := InStateOutputMetricItem{}
	this.LastMessage = lastMessage
	this.Messages = messages
	this.Timestamp = timestamp
	return &this
}

// NewInStateOutputMetricItemWithDefaults instantiates a new InStateOutputMetricItem object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewInStateOutputMetricItemWithDefaults() *InStateOutputMetricItem {
	this := InStateOutputMetricItem{}
	return &this
}

// GetLastMessage returns the LastMessage field value
func (o *InStateOutputMetricItem) GetLastMessage() InStateOutput {
	if o == nil {
		var ret InStateOutput
		return ret
	}

	return o.LastMessage
}

// GetLastMessageOk returns a tuple with the LastMessage field value
// and a boolean to check if the value has been set.
func (o *InStateOutputMetricItem) GetLastMessageOk() (*InStateOutput, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LastMessage, true
}

// SetLastMessage sets field value
func (o *InStateOutputMetricItem) SetLastMessage(v InStateOutput) {
	o.LastMessage = v
}

// GetMessages returns the Messages field value
func (o *InStateOutputMetricItem) GetMessages() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.Messages
}

// GetMessagesOk returns a tuple with the Messages field value
// and a boolean to check if the value has been set.
func (o *InStateOutputMetricItem) GetMessagesOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Messages, true
}

// SetMessages sets field value
func (o *InStateOutputMetricItem) SetMessages(v uint32) {
	o.Messages = v
}

// GetTimestamp returns the Timestamp field value
func (o *InStateOutputMetricItem) GetTimestamp() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.Timestamp
}

// GetTimestampOk returns a tuple with the Timestamp field value
// and a boolean to check if the value has been set.
func (o *InStateOutputMetricItem) GetTimestampOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Timestamp, true
}

// SetTimestamp sets field value
func (o *InStateOutputMetricItem) SetTimestamp(v time.Time) {
	o.Timestamp = v
}

func (o InStateOutputMetricItem) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o InStateOutputMetricItem) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["lastMessage"] = o.LastMessage
	toSerialize["messages"] = o.Messages
	toSerialize["timestamp"] = o.Timestamp
	return toSerialize, nil
}

type NullableInStateOutputMetricItem struct {
	value *InStateOutputMetricItem
	isSet bool
}

func (v NullableInStateOutputMetricItem) Get() *InStateOutputMetricItem {
	return v.value
}

func (v *NullableInStateOutputMetricItem) Set(val *InStateOutputMetricItem) {
	v.value = val
	v.isSet = true
}

func (v NullableInStateOutputMetricItem) IsSet() bool {
	return v.isSet
}

func (v *NullableInStateOutputMetricItem) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInStateOutputMetricItem(val *InStateOutputMetricItem) *NullableInStateOutputMetricItem {
	return &NullableInStateOutputMetricItem{value: val, isSet: true}
}

func (v NullableInStateOutputMetricItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInStateOutputMetricItem) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

