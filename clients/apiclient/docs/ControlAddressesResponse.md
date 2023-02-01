# ControlAddressesResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GoverningAddress** | **string** |  | 
**SinceBlockIndex** | **uint32** |  | 
**StateAddress** | **string** |  | 

## Methods

### NewControlAddressesResponse

`func NewControlAddressesResponse(governingAddress string, sinceBlockIndex uint32, stateAddress string, ) *ControlAddressesResponse`

NewControlAddressesResponse instantiates a new ControlAddressesResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewControlAddressesResponseWithDefaults

`func NewControlAddressesResponseWithDefaults() *ControlAddressesResponse`

NewControlAddressesResponseWithDefaults instantiates a new ControlAddressesResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGoverningAddress

`func (o *ControlAddressesResponse) GetGoverningAddress() string`

GetGoverningAddress returns the GoverningAddress field if non-nil, zero value otherwise.

### GetGoverningAddressOk

`func (o *ControlAddressesResponse) GetGoverningAddressOk() (*string, bool)`

GetGoverningAddressOk returns a tuple with the GoverningAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGoverningAddress

`func (o *ControlAddressesResponse) SetGoverningAddress(v string)`

SetGoverningAddress sets GoverningAddress field to given value.


### GetSinceBlockIndex

`func (o *ControlAddressesResponse) GetSinceBlockIndex() uint32`

GetSinceBlockIndex returns the SinceBlockIndex field if non-nil, zero value otherwise.

### GetSinceBlockIndexOk

`func (o *ControlAddressesResponse) GetSinceBlockIndexOk() (*uint32, bool)`

GetSinceBlockIndexOk returns a tuple with the SinceBlockIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSinceBlockIndex

`func (o *ControlAddressesResponse) SetSinceBlockIndex(v uint32)`

SetSinceBlockIndex sets SinceBlockIndex field to given value.


### GetStateAddress

`func (o *ControlAddressesResponse) GetStateAddress() string`

GetStateAddress returns the StateAddress field if non-nil, zero value otherwise.

### GetStateAddressOk

`func (o *ControlAddressesResponse) GetStateAddressOk() (*string, bool)`

GetStateAddressOk returns a tuple with the StateAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateAddress

`func (o *ControlAddressesResponse) SetStateAddress(v string)`

SetStateAddress sets StateAddress field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


