/*
Wasp API

REST API for the Wasp node

API version: 123
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
	"time"
)

// checks if the TxInclusionStateMsgMetricItem type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &TxInclusionStateMsgMetricItem{}

// TxInclusionStateMsgMetricItem struct for TxInclusionStateMsgMetricItem
type TxInclusionStateMsgMetricItem struct {
	LastMessage TxInclusionStateMsg `json:"lastMessage"`
	Messages uint32 `json:"messages"`
	Timestamp time.Time `json:"timestamp"`
}

// NewTxInclusionStateMsgMetricItem instantiates a new TxInclusionStateMsgMetricItem object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTxInclusionStateMsgMetricItem(lastMessage TxInclusionStateMsg, messages uint32, timestamp time.Time) *TxInclusionStateMsgMetricItem {
	this := TxInclusionStateMsgMetricItem{}
	this.LastMessage = lastMessage
	this.Messages = messages
	this.Timestamp = timestamp
	return &this
}

// NewTxInclusionStateMsgMetricItemWithDefaults instantiates a new TxInclusionStateMsgMetricItem object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTxInclusionStateMsgMetricItemWithDefaults() *TxInclusionStateMsgMetricItem {
	this := TxInclusionStateMsgMetricItem{}
	return &this
}

// GetLastMessage returns the LastMessage field value
func (o *TxInclusionStateMsgMetricItem) GetLastMessage() TxInclusionStateMsg {
	if o == nil {
		var ret TxInclusionStateMsg
		return ret
	}

	return o.LastMessage
}

// GetLastMessageOk returns a tuple with the LastMessage field value
// and a boolean to check if the value has been set.
func (o *TxInclusionStateMsgMetricItem) GetLastMessageOk() (*TxInclusionStateMsg, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LastMessage, true
}

// SetLastMessage sets field value
func (o *TxInclusionStateMsgMetricItem) SetLastMessage(v TxInclusionStateMsg) {
	o.LastMessage = v
}

// GetMessages returns the Messages field value
func (o *TxInclusionStateMsgMetricItem) GetMessages() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.Messages
}

// GetMessagesOk returns a tuple with the Messages field value
// and a boolean to check if the value has been set.
func (o *TxInclusionStateMsgMetricItem) GetMessagesOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Messages, true
}

// SetMessages sets field value
func (o *TxInclusionStateMsgMetricItem) SetMessages(v uint32) {
	o.Messages = v
}

// GetTimestamp returns the Timestamp field value
func (o *TxInclusionStateMsgMetricItem) GetTimestamp() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.Timestamp
}

// GetTimestampOk returns a tuple with the Timestamp field value
// and a boolean to check if the value has been set.
func (o *TxInclusionStateMsgMetricItem) GetTimestampOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Timestamp, true
}

// SetTimestamp sets field value
func (o *TxInclusionStateMsgMetricItem) SetTimestamp(v time.Time) {
	o.Timestamp = v
}

func (o TxInclusionStateMsgMetricItem) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o TxInclusionStateMsgMetricItem) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["lastMessage"] = o.LastMessage
	toSerialize["messages"] = o.Messages
	toSerialize["timestamp"] = o.Timestamp
	return toSerialize, nil
}

type NullableTxInclusionStateMsgMetricItem struct {
	value *TxInclusionStateMsgMetricItem
	isSet bool
}

func (v NullableTxInclusionStateMsgMetricItem) Get() *TxInclusionStateMsgMetricItem {
	return v.value
}

func (v *NullableTxInclusionStateMsgMetricItem) Set(val *TxInclusionStateMsgMetricItem) {
	v.value = val
	v.isSet = true
}

func (v NullableTxInclusionStateMsgMetricItem) IsSet() bool {
	return v.isSet
}

func (v *NullableTxInclusionStateMsgMetricItem) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTxInclusionStateMsgMetricItem(val *TxInclusionStateMsgMetricItem) *NullableTxInclusionStateMsgMetricItem {
	return &NullableTxInclusionStateMsgMetricItem{value: val, isSet: true}
}

func (v NullableTxInclusionStateMsgMetricItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTxInclusionStateMsgMetricItem) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


