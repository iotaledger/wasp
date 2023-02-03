# PeeringTrustedNode

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NetId** | Pointer to **string** | NetID of a peer to trust. | [optional] 
**PubKey** | Pointer to **string** | Public key of the NetID. | [optional] 

## Methods

### NewPeeringTrustedNode

`func NewPeeringTrustedNode() *PeeringTrustedNode`

NewPeeringTrustedNode instantiates a new PeeringTrustedNode object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPeeringTrustedNodeWithDefaults

`func NewPeeringTrustedNodeWithDefaults() *PeeringTrustedNode`

NewPeeringTrustedNodeWithDefaults instantiates a new PeeringTrustedNode object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNetId

`func (o *PeeringTrustedNode) GetNetId() string`

GetNetId returns the NetId field if non-nil, zero value otherwise.

### GetNetIdOk

`func (o *PeeringTrustedNode) GetNetIdOk() (*string, bool)`

GetNetIdOk returns a tuple with the NetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetId

`func (o *PeeringTrustedNode) SetNetId(v string)`

SetNetId sets NetId field to given value.

### HasNetId

`func (o *PeeringTrustedNode) HasNetId() bool`

HasNetId returns a boolean if a field has been set.

### GetPubKey

`func (o *PeeringTrustedNode) GetPubKey() string`

GetPubKey returns the PubKey field if non-nil, zero value otherwise.

### GetPubKeyOk

`func (o *PeeringTrustedNode) GetPubKeyOk() (*string, bool)`

GetPubKeyOk returns a tuple with the PubKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPubKey

`func (o *PeeringTrustedNode) SetPubKey(v string)`

SetPubKey sets PubKey field to given value.

### HasPubKey

`func (o *PeeringTrustedNode) HasPubKey() bool`

HasPubKey returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


