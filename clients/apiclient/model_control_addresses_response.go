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

// checks if the ControlAddressesResponse type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ControlAddressesResponse{}

// ControlAddressesResponse struct for ControlAddressesResponse
type ControlAddressesResponse struct {
	GoverningAddress string `json:"governingAddress"`
	SinceBlockIndex uint32 `json:"sinceBlockIndex"`
	StateAddress string `json:"stateAddress"`
}

// NewControlAddressesResponse instantiates a new ControlAddressesResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewControlAddressesResponse(governingAddress string, sinceBlockIndex uint32, stateAddress string) *ControlAddressesResponse {
	this := ControlAddressesResponse{}
	this.GoverningAddress = governingAddress
	this.SinceBlockIndex = sinceBlockIndex
	this.StateAddress = stateAddress
	return &this
}

// NewControlAddressesResponseWithDefaults instantiates a new ControlAddressesResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewControlAddressesResponseWithDefaults() *ControlAddressesResponse {
	this := ControlAddressesResponse{}
	return &this
}

// GetGoverningAddress returns the GoverningAddress field value
func (o *ControlAddressesResponse) GetGoverningAddress() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.GoverningAddress
}

// GetGoverningAddressOk returns a tuple with the GoverningAddress field value
// and a boolean to check if the value has been set.
func (o *ControlAddressesResponse) GetGoverningAddressOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.GoverningAddress, true
}

// SetGoverningAddress sets field value
func (o *ControlAddressesResponse) SetGoverningAddress(v string) {
	o.GoverningAddress = v
}

// GetSinceBlockIndex returns the SinceBlockIndex field value
func (o *ControlAddressesResponse) GetSinceBlockIndex() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.SinceBlockIndex
}

// GetSinceBlockIndexOk returns a tuple with the SinceBlockIndex field value
// and a boolean to check if the value has been set.
func (o *ControlAddressesResponse) GetSinceBlockIndexOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.SinceBlockIndex, true
}

// SetSinceBlockIndex sets field value
func (o *ControlAddressesResponse) SetSinceBlockIndex(v uint32) {
	o.SinceBlockIndex = v
}

// GetStateAddress returns the StateAddress field value
func (o *ControlAddressesResponse) GetStateAddress() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.StateAddress
}

// GetStateAddressOk returns a tuple with the StateAddress field value
// and a boolean to check if the value has been set.
func (o *ControlAddressesResponse) GetStateAddressOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.StateAddress, true
}

// SetStateAddress sets field value
func (o *ControlAddressesResponse) SetStateAddress(v string) {
	o.StateAddress = v
}

func (o ControlAddressesResponse) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ControlAddressesResponse) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["governingAddress"] = o.GoverningAddress
	toSerialize["sinceBlockIndex"] = o.SinceBlockIndex
	toSerialize["stateAddress"] = o.StateAddress
	return toSerialize, nil
}

type NullableControlAddressesResponse struct {
	value *ControlAddressesResponse
	isSet bool
}

func (v NullableControlAddressesResponse) Get() *ControlAddressesResponse {
	return v.value
}

func (v *NullableControlAddressesResponse) Set(val *ControlAddressesResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableControlAddressesResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableControlAddressesResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableControlAddressesResponse(val *ControlAddressesResponse) *NullableControlAddressesResponse {
	return &NullableControlAddressesResponse{value: val, isSet: true}
}

func (v NullableControlAddressesResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableControlAddressesResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

