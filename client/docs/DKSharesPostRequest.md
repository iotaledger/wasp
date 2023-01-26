# DKSharesPostRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PeerIdentities** | Pointer to **[]string** |  | [optional] 
**Threshold** | Pointer to **int32** | Should be &#x3D;&lt; len(PeerPublicIdentities) | [optional] 
**TimeoutMS** | Pointer to **int32** | Timeout in milliseconds. | [optional] 

## Methods

### NewDKSharesPostRequest

`func NewDKSharesPostRequest() *DKSharesPostRequest`

NewDKSharesPostRequest instantiates a new DKSharesPostRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDKSharesPostRequestWithDefaults

`func NewDKSharesPostRequestWithDefaults() *DKSharesPostRequest`

NewDKSharesPostRequestWithDefaults instantiates a new DKSharesPostRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPeerIdentities

`func (o *DKSharesPostRequest) GetPeerIdentities() []string`

GetPeerIdentities returns the PeerIdentities field if non-nil, zero value otherwise.

### GetPeerIdentitiesOk

`func (o *DKSharesPostRequest) GetPeerIdentitiesOk() (*[]string, bool)`

GetPeerIdentitiesOk returns a tuple with the PeerIdentities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeerIdentities

`func (o *DKSharesPostRequest) SetPeerIdentities(v []string)`

SetPeerIdentities sets PeerIdentities field to given value.

### HasPeerIdentities

`func (o *DKSharesPostRequest) HasPeerIdentities() bool`

HasPeerIdentities returns a boolean if a field has been set.

### GetThreshold

`func (o *DKSharesPostRequest) GetThreshold() int32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *DKSharesPostRequest) GetThresholdOk() (*int32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *DKSharesPostRequest) SetThreshold(v int32)`

SetThreshold sets Threshold field to given value.

### HasThreshold

`func (o *DKSharesPostRequest) HasThreshold() bool`

HasThreshold returns a boolean if a field has been set.

### GetTimeoutMS

`func (o *DKSharesPostRequest) GetTimeoutMS() int32`

GetTimeoutMS returns the TimeoutMS field if non-nil, zero value otherwise.

### GetTimeoutMSOk

`func (o *DKSharesPostRequest) GetTimeoutMSOk() (*int32, bool)`

GetTimeoutMSOk returns a tuple with the TimeoutMS field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeoutMS

`func (o *DKSharesPostRequest) SetTimeoutMS(v int32)`

SetTimeoutMS sets TimeoutMS field to given value.

### HasTimeoutMS

`func (o *DKSharesPostRequest) HasTimeoutMS() bool`

HasTimeoutMS returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


