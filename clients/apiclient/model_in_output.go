/*
Wasp API

REST API for the Wasp node

API version: 0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the InOutput type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &InOutput{}

// InOutput struct for InOutput
type InOutput struct {
	Output Output `json:"output"`
	// The output ID
	OutputId string `json:"outputId"`
}

// NewInOutput instantiates a new InOutput object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewInOutput(output Output, outputId string) *InOutput {
	this := InOutput{}
	this.Output = output
	this.OutputId = outputId
	return &this
}

// NewInOutputWithDefaults instantiates a new InOutput object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewInOutputWithDefaults() *InOutput {
	this := InOutput{}
	return &this
}

// GetOutput returns the Output field value
func (o *InOutput) GetOutput() Output {
	if o == nil {
		var ret Output
		return ret
	}

	return o.Output
}

// GetOutputOk returns a tuple with the Output field value
// and a boolean to check if the value has been set.
func (o *InOutput) GetOutputOk() (*Output, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Output, true
}

// SetOutput sets field value
func (o *InOutput) SetOutput(v Output) {
	o.Output = v
}

// GetOutputId returns the OutputId field value
func (o *InOutput) GetOutputId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.OutputId
}

// GetOutputIdOk returns a tuple with the OutputId field value
// and a boolean to check if the value has been set.
func (o *InOutput) GetOutputIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OutputId, true
}

// SetOutputId sets field value
func (o *InOutput) SetOutputId(v string) {
	o.OutputId = v
}

func (o InOutput) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o InOutput) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["output"] = o.Output
	toSerialize["outputId"] = o.OutputId
	return toSerialize, nil
}

type NullableInOutput struct {
	value *InOutput
	isSet bool
}

func (v NullableInOutput) Get() *InOutput {
	return v.value
}

func (v *NullableInOutput) Set(val *InOutput) {
	v.value = val
	v.isSet = true
}

func (v NullableInOutput) IsSet() bool {
	return v.isSet
}

func (v *NullableInOutput) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInOutput(val *InOutput) *NullableInOutput {
	return &NullableInOutput{value: val, isSet: true}
}

func (v NullableInOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInOutput) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
