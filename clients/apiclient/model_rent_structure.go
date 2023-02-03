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

// checks if the RentStructure type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &RentStructure{}

// RentStructure struct for RentStructure
type RentStructure struct {
	// The virtual byte cost
	VByteCost uint32 `json:"vByteCost"`
	// The virtual byte factor for data fields
	VByteFactorData int32 `json:"vByteFactorData"`
	// The virtual byte factor for key/lookup generating fields
	VByteFactorKey int32 `json:"vByteFactorKey"`
}

// NewRentStructure instantiates a new RentStructure object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRentStructure(vByteCost uint32, vByteFactorData int32, vByteFactorKey int32) *RentStructure {
	this := RentStructure{}
	this.VByteCost = vByteCost
	this.VByteFactorData = vByteFactorData
	this.VByteFactorKey = vByteFactorKey
	return &this
}

// NewRentStructureWithDefaults instantiates a new RentStructure object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRentStructureWithDefaults() *RentStructure {
	this := RentStructure{}
	return &this
}

// GetVByteCost returns the VByteCost field value
func (o *RentStructure) GetVByteCost() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.VByteCost
}

// GetVByteCostOk returns a tuple with the VByteCost field value
// and a boolean to check if the value has been set.
func (o *RentStructure) GetVByteCostOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.VByteCost, true
}

// SetVByteCost sets field value
func (o *RentStructure) SetVByteCost(v uint32) {
	o.VByteCost = v
}

// GetVByteFactorData returns the VByteFactorData field value
func (o *RentStructure) GetVByteFactorData() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.VByteFactorData
}

// GetVByteFactorDataOk returns a tuple with the VByteFactorData field value
// and a boolean to check if the value has been set.
func (o *RentStructure) GetVByteFactorDataOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.VByteFactorData, true
}

// SetVByteFactorData sets field value
func (o *RentStructure) SetVByteFactorData(v int32) {
	o.VByteFactorData = v
}

// GetVByteFactorKey returns the VByteFactorKey field value
func (o *RentStructure) GetVByteFactorKey() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.VByteFactorKey
}

// GetVByteFactorKeyOk returns a tuple with the VByteFactorKey field value
// and a boolean to check if the value has been set.
func (o *RentStructure) GetVByteFactorKeyOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.VByteFactorKey, true
}

// SetVByteFactorKey sets field value
func (o *RentStructure) SetVByteFactorKey(v int32) {
	o.VByteFactorKey = v
}

func (o RentStructure) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o RentStructure) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["vByteCost"] = o.VByteCost
	toSerialize["vByteFactorData"] = o.VByteFactorData
	toSerialize["vByteFactorKey"] = o.VByteFactorKey
	return toSerialize, nil
}

type NullableRentStructure struct {
	value *RentStructure
	isSet bool
}

func (v NullableRentStructure) Get() *RentStructure {
	return v.value
}

func (v *NullableRentStructure) Set(val *RentStructure) {
	v.value = val
	v.isSet = true
}

func (v NullableRentStructure) IsSet() bool {
	return v.isSet
}

func (v *NullableRentStructure) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRentStructure(val *RentStructure) *NullableRentStructure {
	return &NullableRentStructure{value: val, isSet: true}
}

func (v NullableRentStructure) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRentStructure) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


