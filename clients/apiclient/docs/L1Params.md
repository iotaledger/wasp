# L1Params

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BaseToken** | [**BaseToken**](BaseToken.md) |  | 
**MaxPayloadSize** | **uint32** | The max payload size | 
**Protocol** | [**ProtocolParameters**](ProtocolParameters.md) |  | 

## Methods

### NewL1Params

`func NewL1Params(baseToken BaseToken, maxPayloadSize uint32, protocol ProtocolParameters, ) *L1Params`

NewL1Params instantiates a new L1Params object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewL1ParamsWithDefaults

`func NewL1ParamsWithDefaults() *L1Params`

NewL1ParamsWithDefaults instantiates a new L1Params object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBaseToken

`func (o *L1Params) GetBaseToken() BaseToken`

GetBaseToken returns the BaseToken field if non-nil, zero value otherwise.

### GetBaseTokenOk

`func (o *L1Params) GetBaseTokenOk() (*BaseToken, bool)`

GetBaseTokenOk returns a tuple with the BaseToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseToken

`func (o *L1Params) SetBaseToken(v BaseToken)`

SetBaseToken sets BaseToken field to given value.


### GetMaxPayloadSize

`func (o *L1Params) GetMaxPayloadSize() uint32`

GetMaxPayloadSize returns the MaxPayloadSize field if non-nil, zero value otherwise.

### GetMaxPayloadSizeOk

`func (o *L1Params) GetMaxPayloadSizeOk() (*uint32, bool)`

GetMaxPayloadSizeOk returns a tuple with the MaxPayloadSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxPayloadSize

`func (o *L1Params) SetMaxPayloadSize(v uint32)`

SetMaxPayloadSize sets MaxPayloadSize field to given value.


### GetProtocol

`func (o *L1Params) GetProtocol() ProtocolParameters`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *L1Params) GetProtocolOk() (*ProtocolParameters, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *L1Params) SetProtocol(v ProtocolParameters)`

SetProtocol sets Protocol field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


