# PeeringNodeStatusResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsAlive** | **bool** | Whether or not the peer is activated | 
**IsTrusted** | **bool** |  | 
**NetId** | **string** | The NetID of the peer | 
**NumUsers** | **int32** | The amount of users attached to the peer | 
**PublicKey** | **string** | The peers public key encoded in Hex | 

## Methods

### NewPeeringNodeStatusResponse

`func NewPeeringNodeStatusResponse(isAlive bool, isTrusted bool, netId string, numUsers int32, publicKey string, ) *PeeringNodeStatusResponse`

NewPeeringNodeStatusResponse instantiates a new PeeringNodeStatusResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPeeringNodeStatusResponseWithDefaults

`func NewPeeringNodeStatusResponseWithDefaults() *PeeringNodeStatusResponse`

NewPeeringNodeStatusResponseWithDefaults instantiates a new PeeringNodeStatusResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsAlive

`func (o *PeeringNodeStatusResponse) GetIsAlive() bool`

GetIsAlive returns the IsAlive field if non-nil, zero value otherwise.

### GetIsAliveOk

`func (o *PeeringNodeStatusResponse) GetIsAliveOk() (*bool, bool)`

GetIsAliveOk returns a tuple with the IsAlive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsAlive

`func (o *PeeringNodeStatusResponse) SetIsAlive(v bool)`

SetIsAlive sets IsAlive field to given value.


### GetIsTrusted

`func (o *PeeringNodeStatusResponse) GetIsTrusted() bool`

GetIsTrusted returns the IsTrusted field if non-nil, zero value otherwise.

### GetIsTrustedOk

`func (o *PeeringNodeStatusResponse) GetIsTrustedOk() (*bool, bool)`

GetIsTrustedOk returns a tuple with the IsTrusted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsTrusted

`func (o *PeeringNodeStatusResponse) SetIsTrusted(v bool)`

SetIsTrusted sets IsTrusted field to given value.


### GetNetId

`func (o *PeeringNodeStatusResponse) GetNetId() string`

GetNetId returns the NetId field if non-nil, zero value otherwise.

### GetNetIdOk

`func (o *PeeringNodeStatusResponse) GetNetIdOk() (*string, bool)`

GetNetIdOk returns a tuple with the NetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetId

`func (o *PeeringNodeStatusResponse) SetNetId(v string)`

SetNetId sets NetId field to given value.


### GetNumUsers

`func (o *PeeringNodeStatusResponse) GetNumUsers() int32`

GetNumUsers returns the NumUsers field if non-nil, zero value otherwise.

### GetNumUsersOk

`func (o *PeeringNodeStatusResponse) GetNumUsersOk() (*int32, bool)`

GetNumUsersOk returns a tuple with the NumUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumUsers

`func (o *PeeringNodeStatusResponse) SetNumUsers(v int32)`

SetNumUsers sets NumUsers field to given value.


### GetPublicKey

`func (o *PeeringNodeStatusResponse) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *PeeringNodeStatusResponse) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *PeeringNodeStatusResponse) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


