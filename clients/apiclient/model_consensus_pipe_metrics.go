/*
Wasp API

REST API for the Wasp node

API version: 0.4.0-alpha.8-16-g83edf92b9
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the ConsensusPipeMetrics type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ConsensusPipeMetrics{}

// ConsensusPipeMetrics struct for ConsensusPipeMetrics
type ConsensusPipeMetrics struct {
	EventACSMsgPipeSize uint32 `json:"eventACSMsgPipeSize"`
	EventPeerLogIndexMsgPipeSize uint32 `json:"eventPeerLogIndexMsgPipeSize"`
	EventStateTransitionMsgPipeSize uint32 `json:"eventStateTransitionMsgPipeSize"`
	EventTimerMsgPipeSize uint32 `json:"eventTimerMsgPipeSize"`
	EventVMResultMsgPipeSize uint32 `json:"eventVMResultMsgPipeSize"`
}

// NewConsensusPipeMetrics instantiates a new ConsensusPipeMetrics object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConsensusPipeMetrics(eventACSMsgPipeSize uint32, eventPeerLogIndexMsgPipeSize uint32, eventStateTransitionMsgPipeSize uint32, eventTimerMsgPipeSize uint32, eventVMResultMsgPipeSize uint32) *ConsensusPipeMetrics {
	this := ConsensusPipeMetrics{}
	this.EventACSMsgPipeSize = eventACSMsgPipeSize
	this.EventPeerLogIndexMsgPipeSize = eventPeerLogIndexMsgPipeSize
	this.EventStateTransitionMsgPipeSize = eventStateTransitionMsgPipeSize
	this.EventTimerMsgPipeSize = eventTimerMsgPipeSize
	this.EventVMResultMsgPipeSize = eventVMResultMsgPipeSize
	return &this
}

// NewConsensusPipeMetricsWithDefaults instantiates a new ConsensusPipeMetrics object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConsensusPipeMetricsWithDefaults() *ConsensusPipeMetrics {
	this := ConsensusPipeMetrics{}
	return &this
}

// GetEventACSMsgPipeSize returns the EventACSMsgPipeSize field value
func (o *ConsensusPipeMetrics) GetEventACSMsgPipeSize() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.EventACSMsgPipeSize
}

// GetEventACSMsgPipeSizeOk returns a tuple with the EventACSMsgPipeSize field value
// and a boolean to check if the value has been set.
func (o *ConsensusPipeMetrics) GetEventACSMsgPipeSizeOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EventACSMsgPipeSize, true
}

// SetEventACSMsgPipeSize sets field value
func (o *ConsensusPipeMetrics) SetEventACSMsgPipeSize(v uint32) {
	o.EventACSMsgPipeSize = v
}

// GetEventPeerLogIndexMsgPipeSize returns the EventPeerLogIndexMsgPipeSize field value
func (o *ConsensusPipeMetrics) GetEventPeerLogIndexMsgPipeSize() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.EventPeerLogIndexMsgPipeSize
}

// GetEventPeerLogIndexMsgPipeSizeOk returns a tuple with the EventPeerLogIndexMsgPipeSize field value
// and a boolean to check if the value has been set.
func (o *ConsensusPipeMetrics) GetEventPeerLogIndexMsgPipeSizeOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EventPeerLogIndexMsgPipeSize, true
}

// SetEventPeerLogIndexMsgPipeSize sets field value
func (o *ConsensusPipeMetrics) SetEventPeerLogIndexMsgPipeSize(v uint32) {
	o.EventPeerLogIndexMsgPipeSize = v
}

// GetEventStateTransitionMsgPipeSize returns the EventStateTransitionMsgPipeSize field value
func (o *ConsensusPipeMetrics) GetEventStateTransitionMsgPipeSize() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.EventStateTransitionMsgPipeSize
}

// GetEventStateTransitionMsgPipeSizeOk returns a tuple with the EventStateTransitionMsgPipeSize field value
// and a boolean to check if the value has been set.
func (o *ConsensusPipeMetrics) GetEventStateTransitionMsgPipeSizeOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EventStateTransitionMsgPipeSize, true
}

// SetEventStateTransitionMsgPipeSize sets field value
func (o *ConsensusPipeMetrics) SetEventStateTransitionMsgPipeSize(v uint32) {
	o.EventStateTransitionMsgPipeSize = v
}

// GetEventTimerMsgPipeSize returns the EventTimerMsgPipeSize field value
func (o *ConsensusPipeMetrics) GetEventTimerMsgPipeSize() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.EventTimerMsgPipeSize
}

// GetEventTimerMsgPipeSizeOk returns a tuple with the EventTimerMsgPipeSize field value
// and a boolean to check if the value has been set.
func (o *ConsensusPipeMetrics) GetEventTimerMsgPipeSizeOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EventTimerMsgPipeSize, true
}

// SetEventTimerMsgPipeSize sets field value
func (o *ConsensusPipeMetrics) SetEventTimerMsgPipeSize(v uint32) {
	o.EventTimerMsgPipeSize = v
}

// GetEventVMResultMsgPipeSize returns the EventVMResultMsgPipeSize field value
func (o *ConsensusPipeMetrics) GetEventVMResultMsgPipeSize() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.EventVMResultMsgPipeSize
}

// GetEventVMResultMsgPipeSizeOk returns a tuple with the EventVMResultMsgPipeSize field value
// and a boolean to check if the value has been set.
func (o *ConsensusPipeMetrics) GetEventVMResultMsgPipeSizeOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EventVMResultMsgPipeSize, true
}

// SetEventVMResultMsgPipeSize sets field value
func (o *ConsensusPipeMetrics) SetEventVMResultMsgPipeSize(v uint32) {
	o.EventVMResultMsgPipeSize = v
}

func (o ConsensusPipeMetrics) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ConsensusPipeMetrics) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["eventACSMsgPipeSize"] = o.EventACSMsgPipeSize
	toSerialize["eventPeerLogIndexMsgPipeSize"] = o.EventPeerLogIndexMsgPipeSize
	toSerialize["eventStateTransitionMsgPipeSize"] = o.EventStateTransitionMsgPipeSize
	toSerialize["eventTimerMsgPipeSize"] = o.EventTimerMsgPipeSize
	toSerialize["eventVMResultMsgPipeSize"] = o.EventVMResultMsgPipeSize
	return toSerialize, nil
}

type NullableConsensusPipeMetrics struct {
	value *ConsensusPipeMetrics
	isSet bool
}

func (v NullableConsensusPipeMetrics) Get() *ConsensusPipeMetrics {
	return v.value
}

func (v *NullableConsensusPipeMetrics) Set(val *ConsensusPipeMetrics) {
	v.value = val
	v.isSet = true
}

func (v NullableConsensusPipeMetrics) IsSet() bool {
	return v.isSet
}

func (v *NullableConsensusPipeMetrics) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConsensusPipeMetrics(val *ConsensusPipeMetrics) *NullableConsensusPipeMetrics {
	return &NullableConsensusPipeMetrics{value: val, isSet: true}
}

func (v NullableConsensusPipeMetrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConsensusPipeMetrics) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

