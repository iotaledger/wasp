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

// checks if the CallTargetJSON type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CallTargetJSON{}

// CallTargetJSON struct for CallTargetJSON
type CallTargetJSON struct {
	// The contract name as HName (Hex)
	ContractHName string `json:"contractHName"`
	// The function name as HName (Hex)
	FunctionHName string `json:"functionHName"`
}

// NewCallTargetJSON instantiates a new CallTargetJSON object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCallTargetJSON(contractHName string, functionHName string) *CallTargetJSON {
	this := CallTargetJSON{}
	this.ContractHName = contractHName
	this.FunctionHName = functionHName
	return &this
}

// NewCallTargetJSONWithDefaults instantiates a new CallTargetJSON object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCallTargetJSONWithDefaults() *CallTargetJSON {
	this := CallTargetJSON{}
	return &this
}

// GetContractHName returns the ContractHName field value
func (o *CallTargetJSON) GetContractHName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ContractHName
}

// GetContractHNameOk returns a tuple with the ContractHName field value
// and a boolean to check if the value has been set.
func (o *CallTargetJSON) GetContractHNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ContractHName, true
}

// SetContractHName sets field value
func (o *CallTargetJSON) SetContractHName(v string) {
	o.ContractHName = v
}

// GetFunctionHName returns the FunctionHName field value
func (o *CallTargetJSON) GetFunctionHName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.FunctionHName
}

// GetFunctionHNameOk returns a tuple with the FunctionHName field value
// and a boolean to check if the value has been set.
func (o *CallTargetJSON) GetFunctionHNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.FunctionHName, true
}

// SetFunctionHName sets field value
func (o *CallTargetJSON) SetFunctionHName(v string) {
	o.FunctionHName = v
}

func (o CallTargetJSON) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CallTargetJSON) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["contractHName"] = o.ContractHName
	toSerialize["functionHName"] = o.FunctionHName
	return toSerialize, nil
}

type NullableCallTargetJSON struct {
	value *CallTargetJSON
	isSet bool
}

func (v NullableCallTargetJSON) Get() *CallTargetJSON {
	return v.value
}

func (v *NullableCallTargetJSON) Set(val *CallTargetJSON) {
	v.value = val
	v.isSet = true
}

func (v NullableCallTargetJSON) IsSet() bool {
	return v.isSet
}

func (v *NullableCallTargetJSON) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCallTargetJSON(val *CallTargetJSON) *NullableCallTargetJSON {
	return &NullableCallTargetJSON{value: val, isSet: true}
}

func (v NullableCallTargetJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCallTargetJSON) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


