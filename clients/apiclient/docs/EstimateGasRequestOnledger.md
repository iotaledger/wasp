# EstimateGasRequestOnledger

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DryRunRes** | **string** | The result of the DryRun encoded in BCS format and represented in hexadecimal | 
**Msg** | **string** | The ISC Message encoded in BCS format and represented in hexadecimal | 

## Methods

### NewEstimateGasRequestOnledger

`func NewEstimateGasRequestOnledger(dryRunRes string, msg string, ) *EstimateGasRequestOnledger`

NewEstimateGasRequestOnledger instantiates a new EstimateGasRequestOnledger object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEstimateGasRequestOnledgerWithDefaults

`func NewEstimateGasRequestOnledgerWithDefaults() *EstimateGasRequestOnledger`

NewEstimateGasRequestOnledgerWithDefaults instantiates a new EstimateGasRequestOnledger object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDryRunRes

`func (o *EstimateGasRequestOnledger) GetDryRunRes() string`

GetDryRunRes returns the DryRunRes field if non-nil, zero value otherwise.

### GetDryRunResOk

`func (o *EstimateGasRequestOnledger) GetDryRunResOk() (*string, bool)`

GetDryRunResOk returns a tuple with the DryRunRes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDryRunRes

`func (o *EstimateGasRequestOnledger) SetDryRunRes(v string)`

SetDryRunRes sets DryRunRes field to given value.


### GetMsg

`func (o *EstimateGasRequestOnledger) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *EstimateGasRequestOnledger) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *EstimateGasRequestOnledger) SetMsg(v string)`

SetMsg sets Msg field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


