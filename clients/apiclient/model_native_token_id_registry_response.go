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

// checks if the NativeTokenIDRegistryResponse type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &NativeTokenIDRegistryResponse{}

// NativeTokenIDRegistryResponse struct for NativeTokenIDRegistryResponse
type NativeTokenIDRegistryResponse struct {
	NativeTokenRegistryIds []string `json:"nativeTokenRegistryIds"`
}

// NewNativeTokenIDRegistryResponse instantiates a new NativeTokenIDRegistryResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNativeTokenIDRegistryResponse(nativeTokenRegistryIds []string) *NativeTokenIDRegistryResponse {
	this := NativeTokenIDRegistryResponse{}
	this.NativeTokenRegistryIds = nativeTokenRegistryIds
	return &this
}

// NewNativeTokenIDRegistryResponseWithDefaults instantiates a new NativeTokenIDRegistryResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNativeTokenIDRegistryResponseWithDefaults() *NativeTokenIDRegistryResponse {
	this := NativeTokenIDRegistryResponse{}
	return &this
}

// GetNativeTokenRegistryIds returns the NativeTokenRegistryIds field value
func (o *NativeTokenIDRegistryResponse) GetNativeTokenRegistryIds() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.NativeTokenRegistryIds
}

// GetNativeTokenRegistryIdsOk returns a tuple with the NativeTokenRegistryIds field value
// and a boolean to check if the value has been set.
func (o *NativeTokenIDRegistryResponse) GetNativeTokenRegistryIdsOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.NativeTokenRegistryIds, true
}

// SetNativeTokenRegistryIds sets field value
func (o *NativeTokenIDRegistryResponse) SetNativeTokenRegistryIds(v []string) {
	o.NativeTokenRegistryIds = v
}

func (o NativeTokenIDRegistryResponse) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o NativeTokenIDRegistryResponse) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["nativeTokenRegistryIds"] = o.NativeTokenRegistryIds
	return toSerialize, nil
}

type NullableNativeTokenIDRegistryResponse struct {
	value *NativeTokenIDRegistryResponse
	isSet bool
}

func (v NullableNativeTokenIDRegistryResponse) Get() *NativeTokenIDRegistryResponse {
	return v.value
}

func (v *NullableNativeTokenIDRegistryResponse) Set(val *NativeTokenIDRegistryResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableNativeTokenIDRegistryResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableNativeTokenIDRegistryResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNativeTokenIDRegistryResponse(val *NativeTokenIDRegistryResponse) *NullableNativeTokenIDRegistryResponse {
	return &NullableNativeTokenIDRegistryResponse{value: val, isSet: true}
}

func (v NullableNativeTokenIDRegistryResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNativeTokenIDRegistryResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
