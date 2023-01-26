# PeeringNodeStatusResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsAlive** | Pointer to **bool** | Whether or not the peer is activated | [optional] 
**IsTrusted** | Pointer to **bool** |  | [optional] 
**NetId** | Pointer to **string** | The NetID of the peer | [optional] 
**NumUsers** | Pointer to **int32** | The amount of users attached to the peer | [optional] 
**PublicKey** | Pointer to **string** | The peers public key encoded in Hex | [optional] 

## Methods

### NewPeeringNodeStatusResponse

`func NewPeeringNodeStatusResponse() *PeeringNodeStatusResponse`

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

### HasIsAlive

`func (o *PeeringNodeStatusResponse) HasIsAlive() bool`

HasIsAlive returns a boolean if a field has been set.

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

### HasIsTrusted

`func (o *PeeringNodeStatusResponse) HasIsTrusted() bool`

HasIsTrusted returns a boolean if a field has been set.

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

### HasNetId

`func (o *PeeringNodeStatusResponse) HasNetId() bool`

HasNetId returns a boolean if a field has been set.

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

### HasNumUsers

`func (o *PeeringNodeStatusResponse) HasNumUsers() bool`

HasNumUsers returns a boolean if a field has been set.

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

### HasPublicKey

`func (o *PeeringNodeStatusResponse) HasPublicKey() bool`

HasPublicKey returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


