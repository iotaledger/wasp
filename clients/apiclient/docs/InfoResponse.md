# InfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**L1Params** | [**L1Params**](L1Params.md) |  | 
**PeeringURL** | **string** | The net id of the node | 
**PublicKey** | **string** | The public key of the node (Hex) | 
**Version** | **string** | The version of the node | 

## Methods

### NewInfoResponse

`func NewInfoResponse(l1Params L1Params, peeringURL string, publicKey string, version string, ) *InfoResponse`

NewInfoResponse instantiates a new InfoResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInfoResponseWithDefaults

`func NewInfoResponseWithDefaults() *InfoResponse`

NewInfoResponseWithDefaults instantiates a new InfoResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetL1Params

`func (o *InfoResponse) GetL1Params() L1Params`

GetL1Params returns the L1Params field if non-nil, zero value otherwise.

### GetL1ParamsOk

`func (o *InfoResponse) GetL1ParamsOk() (*L1Params, bool)`

GetL1ParamsOk returns a tuple with the L1Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetL1Params

`func (o *InfoResponse) SetL1Params(v L1Params)`

SetL1Params sets L1Params field to given value.


### GetPeeringURL

`func (o *InfoResponse) GetPeeringURL() string`

GetPeeringURL returns the PeeringURL field if non-nil, zero value otherwise.

### GetPeeringURLOk

`func (o *InfoResponse) GetPeeringURLOk() (*string, bool)`

GetPeeringURLOk returns a tuple with the PeeringURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeeringURL

`func (o *InfoResponse) SetPeeringURL(v string)`

SetPeeringURL sets PeeringURL field to given value.


### GetPublicKey

`func (o *InfoResponse) GetPublicKey() string`

GetPublicKey returns the PublicKey field if non-nil, zero value otherwise.

### GetPublicKeyOk

`func (o *InfoResponse) GetPublicKeyOk() (*string, bool)`

GetPublicKeyOk returns a tuple with the PublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicKey

`func (o *InfoResponse) SetPublicKey(v string)`

SetPublicKey sets PublicKey field to given value.


### GetVersion

`func (o *InfoResponse) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *InfoResponse) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *InfoResponse) SetVersion(v string)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


