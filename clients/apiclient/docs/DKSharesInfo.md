# DKSharesInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Address** | **string** | New generated shared address. | 
**PeerIdentities** | **[]string** | Identities of the nodes sharing the key. (Hex) | 
**PeerIndex** | **uint32** |  | 
**PublicKey** | **string** | Used public key. (Hex) | 
**PublicKeyShares** | **[]string** | Public key shares for all the peers. (Hex) | 
**Threshold** | **uint32** |  | 

## Methods

### NewDKSharesInfo

`func NewDKSharesInfo(address string, peerIdentities []string, peerIndex uint32, publicKey string, publicKeyShares []string, threshold uint32, ) *DKSharesInfo`

NewDKSharesInfo instantiates a new DKSharesInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDKSharesInfoWithDefaults

`func NewDKSharesInfoWithDefaults() *DKSharesInfo`

NewDKSharesInfoWithDefaults instantiates a new DKSharesInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAddress

`func (o *DKSharesInfo) GetAddress() string`

GetAddress returns the Address field if non-nil, zero value otherwise.

### GetAddressOk

`func (o *DKSharesInfo) GetAddressOk() (*string, bool)`

GetAddressOk returns a tuple with the Address field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAddress

`func (o *DKSharesInfo) SetAddress(v string)`

SetAddress sets Address field to given value.


### GetPeerIdentities

`func (o *DKSharesInfo) GetPeerIdentities() []string`

GetPeerIdentities returns the PeerIdentities field if non-nil, zero value otherwise.

### GetPeerIdentitiesOk

`func (o *DKSharesInfo) GetPeerIdentitiesOk() (*[]string, bool)`

GetPeerIdentitiesOk returns a tuple with the PeerIdentities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeerIdentities

`func (o *DKSharesInfo) SetPeerIdentities(v []string)`

SetPeerIdentities sets PeerIdentities field to given value.


### GetPeerIndex

`func (o *DKSharesInfo) GetPeerIndex() uint32`

GetPeerIndex returns the PeerIndex field if non-nil, zero value otherwise.

### GetPeerIndexOk

`func (o *DKSharesInfo) GetPeerIndexOk() (*uint32, bool)`

GetPeerIndexOk returns a tuple with the PeerIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeerIndex

`func (o *DKSharesInfo) SetPeerIndex(v uint32)`

SetPeerIndex sets PeerIndex field to given value.


### GetPublicKey

`func (o *DKSharesInfo) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *DKSharesInfo) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *DKSharesInfo) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.


### GetPublicKeyShares

`func (o *DKSharesInfo) GetPublicKeyShares() []string`

GetPublicKeyShares returns the PublicKeyShares field if non-nil, zero value otherwise.

### GetPublicKeySharesOk

`func (o *DKSharesInfo) GetPublicKeySharesOk() (*[]string, bool)`

GetPublicKeySharesOk returns a tuple with the PublicKeyShares field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKeyShares

`func (o *DKSharesInfo) SetPublicKeyShares(v []string)`

SetPublicKeyShares sets PublicKeyShares field to given value.


### GetThreshold

`func (o *DKSharesInfo) GetThreshold() uint32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *DKSharesInfo) GetThresholdOk() (*uint32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *DKSharesInfo) SetThreshold(v uint32)`

SetThreshold sets Threshold field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


