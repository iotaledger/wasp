# NodeOwnerCertificateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OwnerAddress** | **string** | Node owner address. (Bech32) | 
**PublicKey** | **string** | The public key of the node (Hex) | 

## Methods

### NewNodeOwnerCertificateRequest

`func NewNodeOwnerCertificateRequest(ownerAddress string, publicKey string, ) *NodeOwnerCertificateRequest`

NewNodeOwnerCertificateRequest instantiates a new NodeOwnerCertificateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeOwnerCertificateRequestWithDefaults

`func NewNodeOwnerCertificateRequestWithDefaults() *NodeOwnerCertificateRequest`

NewNodeOwnerCertificateRequestWithDefaults instantiates a new NodeOwnerCertificateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOwnerAddress

`func (o *NodeOwnerCertificateRequest) GetOwnerAddress() string`

GetOwnerAddress returns the OwnerAddress field if non-nil, zero value otherwise.

### GetOwnerAddressOk

`func (o *NodeOwnerCertificateRequest) GetOwnerAddressOk() (*string, bool)`

GetOwnerAddressOk returns a tuple with the OwnerAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwnerAddress

`func (o *NodeOwnerCertificateRequest) SetOwnerAddress(v string)`

SetOwnerAddress sets OwnerAddress field to given value.


### GetPublicKey

`func (o *NodeOwnerCertificateRequest) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *NodeOwnerCertificateRequest) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *NodeOwnerCertificateRequest) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


