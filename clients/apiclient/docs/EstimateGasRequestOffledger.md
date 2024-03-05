# EstimateGasRequestOffledger

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FromAddress** | **string** | The address to estimate gas for(Hex) | 
**RequestBytes** | **string** | Offledger Request (Hex) | 

## Methods

### NewEstimateGasRequestOffledger

`func NewEstimateGasRequestOffledger(fromAddress string, requestBytes string, ) *EstimateGasRequestOffledger`

NewEstimateGasRequestOffledger instantiates a new EstimateGasRequestOffledger object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEstimateGasRequestOffledgerWithDefaults

`func NewEstimateGasRequestOffledgerWithDefaults() *EstimateGasRequestOffledger`

NewEstimateGasRequestOffledgerWithDefaults instantiates a new EstimateGasRequestOffledger object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFromAddress

`func (o *EstimateGasRequestOffledger) GetFromAddress() string`

GetFromAddress returns the FromAddress field if non-nil, zero value otherwise.

### GetFromAddressOk

`func (o *EstimateGasRequestOffledger) GetFromAddressOk() (*string, bool)`

GetFromAddressOk returns a tuple with the FromAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFromAddress

`func (o *EstimateGasRequestOffledger) SetFromAddress(v string)`

SetFromAddress sets FromAddress field to given value.


### GetRequestBytes

`func (o *EstimateGasRequestOffledger) GetRequestBytes() string`

GetRequestBytes returns the RequestBytes field if non-nil, zero value otherwise.

### GetRequestBytesOk

`func (o *EstimateGasRequestOffledger) GetRequestBytesOk() (*string, bool)`

GetRequestBytesOk returns a tuple with the RequestBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestBytes

`func (o *EstimateGasRequestOffledger) SetRequestBytes(v string)`

SetRequestBytes sets RequestBytes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


