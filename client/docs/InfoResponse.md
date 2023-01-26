# InfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**L1Params** | Pointer to [**L1Params**](L1Params.md) |  | [optional] 
**NetID** | Pointer to **string** | The net id of the node | [optional] 
**PublicKey** | Pointer to **string** | The public key of the node (Hex) | [optional] 
**Version** | Pointer to **string** | The version of the node | [optional] 

## Methods

### NewInfoResponse

`func NewInfoResponse() *InfoResponse`

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

### HasL1Params

`func (o *InfoResponse) HasL1Params() bool`

HasL1Params returns a boolean if a field has been set.

### GetNetID

`func (o *InfoResponse) GetNetID() string`

GetNetID returns the NetID field if non-nil, zero value otherwise.

### GetNetIDOk

`func (o *InfoResponse) GetNetIDOk() (*string, bool)`

GetNetIDOk returns a tuple with the NetID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetID

`func (o *InfoResponse) SetNetID(v string)`

SetNetID sets NetID field to given value.

### HasNetID

`func (o *InfoResponse) HasNetID() bool`

HasNetID returns a boolean if a field has been set.

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

### HasPublicKey

`func (o *InfoResponse) HasPublicKey() bool`

HasPublicKey returns a boolean if a field has been set.

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

### HasVersion

`func (o *InfoResponse) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


