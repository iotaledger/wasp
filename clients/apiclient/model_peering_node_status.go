/*
Wasp API

REST API for the Wasp node

API version: 123
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the PeeringNodeStatus type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PeeringNodeStatus{}

// PeeringNodeStatus struct for PeeringNodeStatus
type PeeringNodeStatus struct {
	IsAlive *bool `json:"IsAlive,omitempty"`
	NetID *string `json:"NetID,omitempty"`
	NumUsers *uint32 `json:"NumUsers,omitempty"`
	PubKey *string `json:"PubKey,omitempty"`
}

// NewPeeringNodeStatus instantiates a new PeeringNodeStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPeeringNodeStatus() *PeeringNodeStatus {
	this := PeeringNodeStatus{}
	return &this
}

// NewPeeringNodeStatusWithDefaults instantiates a new PeeringNodeStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPeeringNodeStatusWithDefaults() *PeeringNodeStatus {
	this := PeeringNodeStatus{}
	return &this
}

// GetIsAlive returns the IsAlive field value if set, zero value otherwise.
func (o *PeeringNodeStatus) GetIsAlive() bool {
	if o == nil || isNil(o.IsAlive) {
		var ret bool
		return ret
	}
	return *o.IsAlive
}

// GetIsAliveOk returns a tuple with the IsAlive field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PeeringNodeStatus) GetIsAliveOk() (*bool, bool) {
	if o == nil || isNil(o.IsAlive) {
		return nil, false
	}
	return o.IsAlive, true
}

// HasIsAlive returns a boolean if a field has been set.
func (o *PeeringNodeStatus) HasIsAlive() bool {
	if o != nil && !isNil(o.IsAlive) {
		return true
	}

	return false
}

// SetIsAlive gets a reference to the given bool and assigns it to the IsAlive field.
func (o *PeeringNodeStatus) SetIsAlive(v bool) {
	o.IsAlive = &v
}

// GetNetID returns the NetID field value if set, zero value otherwise.
func (o *PeeringNodeStatus) GetNetID() string {
	if o == nil || isNil(o.NetID) {
		var ret string
		return ret
	}
	return *o.NetID
}

// GetNetIDOk returns a tuple with the NetID field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PeeringNodeStatus) GetNetIDOk() (*string, bool) {
	if o == nil || isNil(o.NetID) {
		return nil, false
	}
	return o.NetID, true
}

// HasNetID returns a boolean if a field has been set.
func (o *PeeringNodeStatus) HasNetID() bool {
	if o != nil && !isNil(o.NetID) {
		return true
	}

	return false
}

// SetNetID gets a reference to the given string and assigns it to the NetID field.
func (o *PeeringNodeStatus) SetNetID(v string) {
	o.NetID = &v
}

// GetNumUsers returns the NumUsers field value if set, zero value otherwise.
func (o *PeeringNodeStatus) GetNumUsers() uint32 {
	if o == nil || isNil(o.NumUsers) {
		var ret uint32
		return ret
	}
	return *o.NumUsers
}

// GetNumUsersOk returns a tuple with the NumUsers field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PeeringNodeStatus) GetNumUsersOk() (*uint32, bool) {
	if o == nil || isNil(o.NumUsers) {
		return nil, false
	}
	return o.NumUsers, true
}

// HasNumUsers returns a boolean if a field has been set.
func (o *PeeringNodeStatus) HasNumUsers() bool {
	if o != nil && !isNil(o.NumUsers) {
		return true
	}

	return false
}

// SetNumUsers gets a reference to the given uint32 and assigns it to the NumUsers field.
func (o *PeeringNodeStatus) SetNumUsers(v uint32) {
	o.NumUsers = &v
}

// GetPubKey returns the PubKey field value if set, zero value otherwise.
func (o *PeeringNodeStatus) GetPubKey() string {
	if o == nil || isNil(o.PubKey) {
		var ret string
		return ret
	}
	return *o.PubKey
}

// GetPubKeyOk returns a tuple with the PubKey field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PeeringNodeStatus) GetPubKeyOk() (*string, bool) {
	if o == nil || isNil(o.PubKey) {
		return nil, false
	}
	return o.PubKey, true
}

// HasPubKey returns a boolean if a field has been set.
func (o *PeeringNodeStatus) HasPubKey() bool {
	if o != nil && !isNil(o.PubKey) {
		return true
	}

	return false
}

// SetPubKey gets a reference to the given string and assigns it to the PubKey field.
func (o *PeeringNodeStatus) SetPubKey(v string) {
	o.PubKey = &v
}

func (o PeeringNodeStatus) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PeeringNodeStatus) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.IsAlive) {
		toSerialize["IsAlive"] = o.IsAlive
	}
	if !isNil(o.NetID) {
		toSerialize["NetID"] = o.NetID
	}
	if !isNil(o.NumUsers) {
		toSerialize["NumUsers"] = o.NumUsers
	}
	if !isNil(o.PubKey) {
		toSerialize["PubKey"] = o.PubKey
	}
	return toSerialize, nil
}

type NullablePeeringNodeStatus struct {
	value *PeeringNodeStatus
	isSet bool
}

func (v NullablePeeringNodeStatus) Get() *PeeringNodeStatus {
	return v.value
}

func (v *NullablePeeringNodeStatus) Set(val *PeeringNodeStatus) {
	v.value = val
	v.isSet = true
}

func (v NullablePeeringNodeStatus) IsSet() bool {
	return v.isSet
}

func (v *NullablePeeringNodeStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePeeringNodeStatus(val *PeeringNodeStatus) *NullablePeeringNodeStatus {
	return &NullablePeeringNodeStatus{value: val, isSet: true}
}

func (v NullablePeeringNodeStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePeeringNodeStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

