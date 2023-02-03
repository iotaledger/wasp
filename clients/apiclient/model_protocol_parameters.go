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

// checks if the ProtocolParameters type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ProtocolParameters{}

// ProtocolParameters struct for ProtocolParameters
type ProtocolParameters struct {
	// The human readable network prefix
	Bech32Hrp string `json:"bech32Hrp"`
	// The networks max depth
	BelowMaxDepth uint32 `json:"belowMaxDepth"`
	// The minimal PoW score
	MinPowScore uint32 `json:"minPowScore"`
	// The network name
	NetworkName string `json:"networkName"`
	RentStructure RentStructure `json:"rentStructure"`
	// The token supply
	TokenSupply string `json:"tokenSupply"`
	// The protocol version
	Version int32 `json:"version"`
}

// NewProtocolParameters instantiates a new ProtocolParameters object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtocolParameters(bech32Hrp string, belowMaxDepth uint32, minPowScore uint32, networkName string, rentStructure RentStructure, tokenSupply string, version int32) *ProtocolParameters {
	this := ProtocolParameters{}
	this.Bech32Hrp = bech32Hrp
	this.BelowMaxDepth = belowMaxDepth
	this.MinPowScore = minPowScore
	this.NetworkName = networkName
	this.RentStructure = rentStructure
	this.TokenSupply = tokenSupply
	this.Version = version
	return &this
}

// NewProtocolParametersWithDefaults instantiates a new ProtocolParameters object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtocolParametersWithDefaults() *ProtocolParameters {
	this := ProtocolParameters{}
	return &this
}

// GetBech32Hrp returns the Bech32Hrp field value
func (o *ProtocolParameters) GetBech32Hrp() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Bech32Hrp
}

// GetBech32HrpOk returns a tuple with the Bech32Hrp field value
// and a boolean to check if the value has been set.
func (o *ProtocolParameters) GetBech32HrpOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Bech32Hrp, true
}

// SetBech32Hrp sets field value
func (o *ProtocolParameters) SetBech32Hrp(v string) {
	o.Bech32Hrp = v
}

// GetBelowMaxDepth returns the BelowMaxDepth field value
func (o *ProtocolParameters) GetBelowMaxDepth() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.BelowMaxDepth
}

// GetBelowMaxDepthOk returns a tuple with the BelowMaxDepth field value
// and a boolean to check if the value has been set.
func (o *ProtocolParameters) GetBelowMaxDepthOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.BelowMaxDepth, true
}

// SetBelowMaxDepth sets field value
func (o *ProtocolParameters) SetBelowMaxDepth(v uint32) {
	o.BelowMaxDepth = v
}

// GetMinPowScore returns the MinPowScore field value
func (o *ProtocolParameters) GetMinPowScore() uint32 {
	if o == nil {
		var ret uint32
		return ret
	}

	return o.MinPowScore
}

// GetMinPowScoreOk returns a tuple with the MinPowScore field value
// and a boolean to check if the value has been set.
func (o *ProtocolParameters) GetMinPowScoreOk() (*uint32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.MinPowScore, true
}

// SetMinPowScore sets field value
func (o *ProtocolParameters) SetMinPowScore(v uint32) {
	o.MinPowScore = v
}

// GetNetworkName returns the NetworkName field value
func (o *ProtocolParameters) GetNetworkName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.NetworkName
}

// GetNetworkNameOk returns a tuple with the NetworkName field value
// and a boolean to check if the value has been set.
func (o *ProtocolParameters) GetNetworkNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NetworkName, true
}

// SetNetworkName sets field value
func (o *ProtocolParameters) SetNetworkName(v string) {
	o.NetworkName = v
}

// GetRentStructure returns the RentStructure field value
func (o *ProtocolParameters) GetRentStructure() RentStructure {
	if o == nil {
		var ret RentStructure
		return ret
	}

	return o.RentStructure
}

// GetRentStructureOk returns a tuple with the RentStructure field value
// and a boolean to check if the value has been set.
func (o *ProtocolParameters) GetRentStructureOk() (*RentStructure, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RentStructure, true
}

// SetRentStructure sets field value
func (o *ProtocolParameters) SetRentStructure(v RentStructure) {
	o.RentStructure = v
}

// GetTokenSupply returns the TokenSupply field value
func (o *ProtocolParameters) GetTokenSupply() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TokenSupply
}

// GetTokenSupplyOk returns a tuple with the TokenSupply field value
// and a boolean to check if the value has been set.
func (o *ProtocolParameters) GetTokenSupplyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TokenSupply, true
}

// SetTokenSupply sets field value
func (o *ProtocolParameters) SetTokenSupply(v string) {
	o.TokenSupply = v
}

// GetVersion returns the Version field value
func (o *ProtocolParameters) GetVersion() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Version
}

// GetVersionOk returns a tuple with the Version field value
// and a boolean to check if the value has been set.
func (o *ProtocolParameters) GetVersionOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Version, true
}

// SetVersion sets field value
func (o *ProtocolParameters) SetVersion(v int32) {
	o.Version = v
}

func (o ProtocolParameters) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ProtocolParameters) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["bech32Hrp"] = o.Bech32Hrp
	toSerialize["belowMaxDepth"] = o.BelowMaxDepth
	toSerialize["minPowScore"] = o.MinPowScore
	toSerialize["networkName"] = o.NetworkName
	toSerialize["rentStructure"] = o.RentStructure
	toSerialize["tokenSupply"] = o.TokenSupply
	toSerialize["version"] = o.Version
	return toSerialize, nil
}

type NullableProtocolParameters struct {
	value *ProtocolParameters
	isSet bool
}

func (v NullableProtocolParameters) Get() *ProtocolParameters {
	return v.value
}

func (v *NullableProtocolParameters) Set(val *ProtocolParameters) {
	v.value = val
	v.isSet = true
}

func (v NullableProtocolParameters) IsSet() bool {
	return v.isSet
}

func (v *NullableProtocolParameters) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtocolParameters(val *ProtocolParameters) *NullableProtocolParameters {
	return &NullableProtocolParameters{value: val, isSet: true}
}

func (v NullableProtocolParameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtocolParameters) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


