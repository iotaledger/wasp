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

// checks if the DKSharesPostRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &DKSharesPostRequest{}

// DKSharesPostRequest struct for DKSharesPostRequest
type DKSharesPostRequest struct {
	PeerIdentities []string `json:"peerIdentities"`
	// Should be =< len(PeerPublicIdentities)
	Threshold uint32 `json:"threshold"`
	// Timeout in milliseconds.
	TimeoutMS uint32 `json:"timeoutMS"`
}

// NewDKSharesPostRequest instantiates a new DKSharesPostRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDKSharesPostRequest(peerIdentities []string, threshold uint32, timeoutMS uint32) *DKSharesPostRequest {
	this := DKSharesPostRequest{}
	this.PeerIdentities = peerIdentities
	this.Threshold = threshold
	this.TimeoutMS = timeoutMS
	return &this
}

// NewDKSharesPostRequestWithDefaults instantiates a new DKSharesPostRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDKSharesPostRequestWithDefaults() *DKSharesPostRequest {
	this := DKSharesPostRequest{}
	return &this
}

// GetPeerIdentities returns the PeerIdentities field value
func (o *DKSharesPostRequest) GetPeerIdentities() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.PeerIdentities
}

// GetPeerIdentitiesOk returns a tuple with the PeerIdentities field value
// and a boolean to check if the value has been set.
func (o *DKSharesPostRequest) GetPeerIdentitiesOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.PeerIdentities, true
}

// SetPeerIdentities sets field value
func (o *DKSharesPostRequest) SetPeerIdentities(v []string) {
	o.PeerIdentities = v
}

// GetThreshold returns the Threshold field value
func (o *DKSharesPostRequest) GetThreshold() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.Threshold
}

// GetThresholdOk returns a tuple with the Threshold field value
// and a boolean to check if the value has been set.
func (o *DKSharesPostRequest) GetThresholdOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Threshold, true
}

// SetThreshold sets field value
func (o *DKSharesPostRequest) SetThreshold(v uint32) {
	o.Threshold = v
}

// GetTimeoutMS returns the TimeoutMS field value
func (o *DKSharesPostRequest) GetTimeoutMS() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.TimeoutMS
}

// GetTimeoutMSOk returns a tuple with the TimeoutMS field value
// and a boolean to check if the value has been set.
func (o *DKSharesPostRequest) GetTimeoutMSOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TimeoutMS, true
}

// SetTimeoutMS sets field value
func (o *DKSharesPostRequest) SetTimeoutMS(v uint32) {
	o.TimeoutMS = v
}

func (o DKSharesPostRequest) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o DKSharesPostRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["peerIdentities"] = o.PeerIdentities
	toSerialize["threshold"] = o.Threshold
	toSerialize["timeoutMS"] = o.TimeoutMS
	return toSerialize, nil
}

type NullableDKSharesPostRequest struct {
	value *DKSharesPostRequest
	isSet bool
}

func (v NullableDKSharesPostRequest) Get() *DKSharesPostRequest {
	return v.value
}

func (v *NullableDKSharesPostRequest) Set(val *DKSharesPostRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableDKSharesPostRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableDKSharesPostRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDKSharesPostRequest(val *DKSharesPostRequest) *NullableDKSharesPostRequest {
	return &NullableDKSharesPostRequest{value: val, isSet: true}
}

func (v NullableDKSharesPostRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDKSharesPostRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


