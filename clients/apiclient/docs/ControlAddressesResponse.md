# ControlAddressesResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AnchorOwner** | **string** | The anchor owner (Hex Address) | 
**ChainAdmin** | **string** | The chain admin (Hex Address) | 
**SinceBlockIndex** | **int32** | The block index (uint32 | 

## Methods

### NewControlAddressesResponse

`func NewControlAddressesResponse(anchorOwner string, chainAdmin string, sinceBlockIndex int32, ) *ControlAddressesResponse`

NewControlAddressesResponse instantiates a new ControlAddressesResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewControlAddressesResponseWithDefaults

`func NewControlAddressesResponseWithDefaults() *ControlAddressesResponse`

NewControlAddressesResponseWithDefaults instantiates a new ControlAddressesResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnchorOwner

`func (o *ControlAddressesResponse) GetAnchorOwner() string`

GetAnchorOwner returns the AnchorOwner field if non-nil, zero value otherwise.

### GetAnchorOwnerOk

`func (o *ControlAddressesResponse) GetAnchorOwnerOk() (*string, bool)`

GetAnchorOwnerOk returns a tuple with the AnchorOwner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnchorOwner

`func (o *ControlAddressesResponse) SetAnchorOwner(v string)`

SetAnchorOwner sets AnchorOwner field to given value.


### GetChainAdmin

`func (o *ControlAddressesResponse) GetChainAdmin() string`

GetChainAdmin returns the ChainAdmin field if non-nil, zero value otherwise.

### GetChainAdminOk

`func (o *ControlAddressesResponse) GetChainAdminOk() (*string, bool)`

GetChainAdminOk returns a tuple with the ChainAdmin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainAdmin

`func (o *ControlAddressesResponse) SetChainAdmin(v string)`

SetChainAdmin sets ChainAdmin field to given value.


### GetSinceBlockIndex

`func (o *ControlAddressesResponse) GetSinceBlockIndex() int32`

GetSinceBlockIndex returns the SinceBlockIndex field if non-nil, zero value otherwise.

### GetSinceBlockIndexOk

`func (o *ControlAddressesResponse) GetSinceBlockIndexOk() (*int32, bool)`

GetSinceBlockIndexOk returns a tuple with the SinceBlockIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSinceBlockIndex

`func (o *ControlAddressesResponse) SetSinceBlockIndex(v int32)`

SetSinceBlockIndex sets SinceBlockIndex field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


