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

// checks if the Assets type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Assets{}

// Assets struct for Assets
type Assets struct {
	BaseTokens int64 `json:"baseTokens"`
	NativeTokens []NativeToken `json:"nativeTokens"`
	Nfts []string `json:"nfts"`
}

// NewAssets instantiates a new Assets object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAssets(baseTokens int64, nativeTokens []NativeToken, nfts []string) *Assets {
	this := Assets{}
	this.BaseTokens = baseTokens
	this.NativeTokens = nativeTokens
	this.Nfts = nfts
	return &this
}

// NewAssetsWithDefaults instantiates a new Assets object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAssetsWithDefaults() *Assets {
	this := Assets{}
	return &this
}

// GetBaseTokens returns the BaseTokens field value
func (o *Assets) GetBaseTokens() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.BaseTokens
}

// GetBaseTokensOk returns a tuple with the BaseTokens field value
// and a boolean to check if the value has been set.
func (o *Assets) GetBaseTokensOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.BaseTokens, true
}

// SetBaseTokens sets field value
func (o *Assets) SetBaseTokens(v int64) {
	o.BaseTokens = v
}

// GetNativeTokens returns the NativeTokens field value
func (o *Assets) GetNativeTokens() []NativeToken {
	if o == nil {
		var ret []NativeToken
		return ret
	}

	return o.NativeTokens
}

// GetNativeTokensOk returns a tuple with the NativeTokens field value
// and a boolean to check if the value has been set.
func (o *Assets) GetNativeTokensOk() ([]NativeToken, bool) {
	if o == nil {
		return nil, false
	}
	return o.NativeTokens, true
}

// SetNativeTokens sets field value
func (o *Assets) SetNativeTokens(v []NativeToken) {
	o.NativeTokens = v
}

// GetNfts returns the Nfts field value
func (o *Assets) GetNfts() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Nfts
}

// GetNftsOk returns a tuple with the Nfts field value
// and a boolean to check if the value has been set.
func (o *Assets) GetNftsOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Nfts, true
}

// SetNfts sets field value
func (o *Assets) SetNfts(v []string) {
	o.Nfts = v
}

func (o Assets) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Assets) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["baseTokens"] = o.BaseTokens
	toSerialize["nativeTokens"] = o.NativeTokens
	toSerialize["nfts"] = o.Nfts
	return toSerialize, nil
}

type NullableAssets struct {
	value *Assets
	isSet bool
}

func (v NullableAssets) Get() *Assets {
	return v.value
}

func (v *NullableAssets) Set(val *Assets) {
	v.value = val
	v.isSet = true
}

func (v NullableAssets) IsSet() bool {
	return v.isSet
}

func (v *NullableAssets) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAssets(val *Assets) *NullableAssets {
	return &NullableAssets{value: val, isSet: true}
}

func (v NullableAssets) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAssets) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

