# PeeringNodeIdentityResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsTrusted** | **bool** |  | 
**NetId** | **string** | The NetID of the peer | 
**PublicKey** | **string** | The peers public key encoded in Hex | 

## Methods

### NewPeeringNodeIdentityResponse

`func NewPeeringNodeIdentityResponse(isTrusted bool, netId string, publicKey string, ) *PeeringNodeIdentityResponse`

NewPeeringNodeIdentityResponse instantiates a new PeeringNodeIdentityResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPeeringNodeIdentityResponseWithDefaults

`func NewPeeringNodeIdentityResponseWithDefaults() *PeeringNodeIdentityResponse`

NewPeeringNodeIdentityResponseWithDefaults instantiates a new PeeringNodeIdentityResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsTrusted

`func (o *PeeringNodeIdentityResponse) GetIsTrusted() bool`

GetIsTrusted returns the IsTrusted field if non-nil, zero value otherwise.

### GetIsTrustedOk

`func (o *PeeringNodeIdentityResponse) GetIsTrustedOk() (*bool, bool)`

GetIsTrustedOk returns a tuple with the IsTrusted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsTrusted

`func (o *PeeringNodeIdentityResponse) SetIsTrusted(v bool)`

SetIsTrusted sets IsTrusted field to given value.


### GetNetId

`func (o *PeeringNodeIdentityResponse) GetNetId() string`

GetNetId returns the NetId field if non-nil, zero value otherwise.

### GetNetIdOk

`func (o *PeeringNodeIdentityResponse) GetNetIdOk() (*string, bool)`

GetNetIdOk returns a tuple with the NetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetId

`func (o *PeeringNodeIdentityResponse) SetNetId(v string)`

SetNetId sets NetId field to given value.


### GetPublicKey

`func (o *PeeringNodeIdentityResponse) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *PeeringNodeIdentityResponse) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *PeeringNodeIdentityResponse) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


