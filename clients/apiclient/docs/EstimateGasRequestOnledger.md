# EstimateGasRequestOnledger

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TransactionBytes** | **string** | The result of the DryRun encoded in BCS format and represented in hexadecimal | 

## Methods

### NewEstimateGasRequestOnledger

`func NewEstimateGasRequestOnledger(transactionBytes string, ) *EstimateGasRequestOnledger`

NewEstimateGasRequestOnledger instantiates a new EstimateGasRequestOnledger object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEstimateGasRequestOnledgerWithDefaults

`func NewEstimateGasRequestOnledgerWithDefaults() *EstimateGasRequestOnledger`

NewEstimateGasRequestOnledgerWithDefaults instantiates a new EstimateGasRequestOnledger object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTransactionBytes

`func (o *EstimateGasRequestOnledger) GetTransactionBytes() string`

GetTransactionBytes returns the TransactionBytes field if non-nil, zero value otherwise.

### GetTransactionBytesOk

`func (o *EstimateGasRequestOnledger) GetTransactionBytesOk() (*string, bool)`

GetTransactionBytesOk returns a tuple with the TransactionBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionBytes

`func (o *EstimateGasRequestOnledger) SetTransactionBytes(v string)`

SetTransactionBytes sets TransactionBytes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


