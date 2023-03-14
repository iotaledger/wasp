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

// checks if the NodeOwnerCertificateResponse type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &NodeOwnerCertificateResponse{}

// NodeOwnerCertificateResponse struct for NodeOwnerCertificateResponse
type NodeOwnerCertificateResponse struct {
	// Certificate stating the ownership. (Hex)
	Certificate string `json:"certificate"`
}

// NewNodeOwnerCertificateResponse instantiates a new NodeOwnerCertificateResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNodeOwnerCertificateResponse(certificate string) *NodeOwnerCertificateResponse {
	this := NodeOwnerCertificateResponse{}
	this.Certificate = certificate
	return &this
}

// NewNodeOwnerCertificateResponseWithDefaults instantiates a new NodeOwnerCertificateResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNodeOwnerCertificateResponseWithDefaults() *NodeOwnerCertificateResponse {
	this := NodeOwnerCertificateResponse{}
	return &this
}

// GetCertificate returns the Certificate field value
func (o *NodeOwnerCertificateResponse) GetCertificate() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Certificate
}

// GetCertificateOk returns a tuple with the Certificate field value
// and a boolean to check if the value has been set.
func (o *NodeOwnerCertificateResponse) GetCertificateOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Certificate, true
}

// SetCertificate sets field value
func (o *NodeOwnerCertificateResponse) SetCertificate(v string) {
	o.Certificate = v
}

func (o NodeOwnerCertificateResponse) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o NodeOwnerCertificateResponse) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["certificate"] = o.Certificate
	return toSerialize, nil
}

type NullableNodeOwnerCertificateResponse struct {
	value *NodeOwnerCertificateResponse
	isSet bool
}

func (v NullableNodeOwnerCertificateResponse) Get() *NodeOwnerCertificateResponse {
	return v.value
}

func (v *NullableNodeOwnerCertificateResponse) Set(val *NodeOwnerCertificateResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableNodeOwnerCertificateResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableNodeOwnerCertificateResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNodeOwnerCertificateResponse(val *NodeOwnerCertificateResponse) *NullableNodeOwnerCertificateResponse {
	return &NullableNodeOwnerCertificateResponse{value: val, isSet: true}
}

func (v NullableNodeOwnerCertificateResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNodeOwnerCertificateResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
