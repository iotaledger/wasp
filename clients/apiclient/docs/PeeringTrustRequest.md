# PeeringTrustRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**PeeringURL** | **string** | The peering URL of the peer | 
**PublicKey** | **string** | The peers public key encoded in Hex | 

## Methods

### NewPeeringTrustRequest

`func NewPeeringTrustRequest(name string, peeringURL string, publicKey string, ) *PeeringTrustRequest`

NewPeeringTrustRequest instantiates a new PeeringTrustRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPeeringTrustRequestWithDefaults

`func NewPeeringTrustRequestWithDefaults() *PeeringTrustRequest`

NewPeeringTrustRequestWithDefaults instantiates a new PeeringTrustRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PeeringTrustRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PeeringTrustRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PeeringTrustRequest) SetName(v string)`

SetName sets Name field to given value.


### GetPeeringURL

`func (o *PeeringTrustRequest) GetPeeringURL() string`

GetPeeringURL returns the PeeringURL field if non-nil, zero value otherwise.

### GetPeeringURLOk

`func (o *PeeringTrustRequest) GetPeeringURLOk() (*string, bool)`

GetPeeringURLOk returns a tuple with the PeeringURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeeringURL

`func (o *PeeringTrustRequest) SetPeeringURL(v string)`

SetPeeringURL sets PeeringURL field to given value.


### GetPublicKey

`func (o *PeeringTrustRequest) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *PeeringTrustRequest) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *PeeringTrustRequest) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


