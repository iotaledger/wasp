/*
Wasp API

REST API for the Wasp node

API version: 0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the ContractInfoResponse type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ContractInfoResponse{}

// ContractInfoResponse struct for ContractInfoResponse
type ContractInfoResponse struct {
	// The id (HName as Hex)) of the contract.
	HName string `json:"hName"`
	// The name of the contract.
	Name string `json:"name"`
}

type _ContractInfoResponse ContractInfoResponse

// NewContractInfoResponse instantiates a new ContractInfoResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewContractInfoResponse(hName string, name string) *ContractInfoResponse {
	this := ContractInfoResponse{}
	this.HName = hName
	this.Name = name
	return &this
}

// NewContractInfoResponseWithDefaults instantiates a new ContractInfoResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewContractInfoResponseWithDefaults() *ContractInfoResponse {
	this := ContractInfoResponse{}
	return &this
}

// GetHName returns the HName field value
func (o *ContractInfoResponse) GetHName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.HName
}

// GetHNameOk returns a tuple with the HName field value
// and a boolean to check if the value has been set.
func (o *ContractInfoResponse) GetHNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.HName, true
}

// SetHName sets field value
func (o *ContractInfoResponse) SetHName(v string) {
	o.HName = v
}

// GetName returns the Name field value
func (o *ContractInfoResponse) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *ContractInfoResponse) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *ContractInfoResponse) SetName(v string) {
	o.Name = v
}

func (o ContractInfoResponse) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ContractInfoResponse) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["hName"] = o.HName
	toSerialize["name"] = o.Name
	return toSerialize, nil
}

func (o *ContractInfoResponse) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"hName",
		"name",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varContractInfoResponse := _ContractInfoResponse{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varContractInfoResponse)

	if err != nil {
		return err
	}

	*o = ContractInfoResponse(varContractInfoResponse)

	return err
}

type NullableContractInfoResponse struct {
	value *ContractInfoResponse
	isSet bool
}

func (v NullableContractInfoResponse) Get() *ContractInfoResponse {
	return v.value
}

func (v *NullableContractInfoResponse) Set(val *ContractInfoResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableContractInfoResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableContractInfoResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableContractInfoResponse(val *ContractInfoResponse) *NullableContractInfoResponse {
	return &NullableContractInfoResponse{value: val, isSet: true}
}

func (v NullableContractInfoResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableContractInfoResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


