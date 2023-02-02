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

// checks if the BlockReceiptsResponse type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &BlockReceiptsResponse{}

// BlockReceiptsResponse struct for BlockReceiptsResponse
type BlockReceiptsResponse struct {
	Receipts []RequestReceiptResponse `json:"receipts"`
}

// NewBlockReceiptsResponse instantiates a new BlockReceiptsResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewBlockReceiptsResponse(receipts []RequestReceiptResponse) *BlockReceiptsResponse {
	this := BlockReceiptsResponse{}
	this.Receipts = receipts
	return &this
}

// NewBlockReceiptsResponseWithDefaults instantiates a new BlockReceiptsResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewBlockReceiptsResponseWithDefaults() *BlockReceiptsResponse {
	this := BlockReceiptsResponse{}
	return &this
}

// GetReceipts returns the Receipts field value
func (o *BlockReceiptsResponse) GetReceipts() []RequestReceiptResponse {
	if o == nil {
		var ret []RequestReceiptResponse
		return ret
	}

	return o.Receipts
}

// GetReceiptsOk returns a tuple with the Receipts field value
// and a boolean to check if the value has been set.
func (o *BlockReceiptsResponse) GetReceiptsOk() ([]RequestReceiptResponse, bool) {
	if o == nil {
		return nil, false
	}
	return o.Receipts, true
}

// SetReceipts sets field value
func (o *BlockReceiptsResponse) SetReceipts(v []RequestReceiptResponse) {
	o.Receipts = v
}

func (o BlockReceiptsResponse) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o BlockReceiptsResponse) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["receipts"] = o.Receipts
	return toSerialize, nil
}

type NullableBlockReceiptsResponse struct {
	value *BlockReceiptsResponse
	isSet bool
}

func (v NullableBlockReceiptsResponse) Get() *BlockReceiptsResponse {
	return v.value
}

func (v *NullableBlockReceiptsResponse) Set(val *BlockReceiptsResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableBlockReceiptsResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableBlockReceiptsResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableBlockReceiptsResponse(val *BlockReceiptsResponse) *NullableBlockReceiptsResponse {
	return &NullableBlockReceiptsResponse{value: val, isSet: true}
}

func (v NullableBlockReceiptsResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableBlockReceiptsResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


