# PeeringNodeStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsAlive** | Pointer to **bool** |  | [optional] 
**NetID** | Pointer to **string** |  | [optional] 
**NumUsers** | Pointer to **uint32** |  | [optional] 
**PubKey** | Pointer to **string** |  | [optional] 

## Methods

### NewPeeringNodeStatus

`func NewPeeringNodeStatus() *PeeringNodeStatus`

NewPeeringNodeStatus instantiates a new PeeringNodeStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPeeringNodeStatusWithDefaults

`func NewPeeringNodeStatusWithDefaults() *PeeringNodeStatus`

NewPeeringNodeStatusWithDefaults instantiates a new PeeringNodeStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsAlive

`func (o *PeeringNodeStatus) GetIsAlive() bool`

GetIsAlive returns the IsAlive field if non-nil, zero value otherwise.

### GetIsAliveOk

`func (o *PeeringNodeStatus) GetIsAliveOk() (*bool, bool)`

GetIsAliveOk returns a tuple with the IsAlive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsAlive

`func (o *PeeringNodeStatus) SetIsAlive(v bool)`

SetIsAlive sets IsAlive field to given value.

### HasIsAlive

`func (o *PeeringNodeStatus) HasIsAlive() bool`

HasIsAlive returns a boolean if a field has been set.

### GetNetID

`func (o *PeeringNodeStatus) GetNetID() string`

GetNetID returns the NetID field if non-nil, zero value otherwise.

### GetNetIDOk

`func (o *PeeringNodeStatus) GetNetIDOk() (*string, bool)`

GetNetIDOk returns a tuple with the NetID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetID

`func (o *PeeringNodeStatus) SetNetID(v string)`

SetNetID sets NetID field to given value.

### HasNetID

`func (o *PeeringNodeStatus) HasNetID() bool`

HasNetID returns a boolean if a field has been set.

### GetNumUsers

`func (o *PeeringNodeStatus) GetNumUsers() uint32`

GetNumUsers returns the NumUsers field if non-nil, zero value otherwise.

### GetNumUsersOk

`func (o *PeeringNodeStatus) GetNumUsersOk() (*uint32, bool)`

GetNumUsersOk returns a tuple with the NumUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumUsers

`func (o *PeeringNodeStatus) SetNumUsers(v uint32)`

SetNumUsers sets NumUsers field to given value.

### HasNumUsers

`func (o *PeeringNodeStatus) HasNumUsers() bool`

HasNumUsers returns a boolean if a field has been set.

### GetPubKey

`func (o *PeeringNodeStatus) GetPubKey() string`

GetPubKey returns the PubKey field if non-nil, zero value otherwise.

### GetPubKeyOk

`func (o *PeeringNodeStatus) GetPubKeyOk() (*string, bool)`

GetPubKeyOk returns a tuple with the PubKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPubKey

`func (o *PeeringNodeStatus) SetPubKey(v string)`

SetPubKey sets PubKey field to given value.

### HasPubKey

`func (o *PeeringNodeStatus) HasPubKey() bool`

HasPubKey returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


