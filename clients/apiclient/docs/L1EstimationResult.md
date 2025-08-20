# L1EstimationResult

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ComputationFee** | **string** | Gas cost for computation (uint64 as string) | 
**GasBudget** | **string** |  | 
**GasFeeCharged** | **string** | Total gas fee charged: computation fee + storage fee - storage rebate (uint64 as string) | 
**StorageFee** | **string** | Gas cost for storage (uint64 as string) | 
**StorageRebate** | **string** | Gas rebate for storage (uint64 as string) | 

## Methods

### NewL1EstimationResult

`func NewL1EstimationResult(computationFee string, gasBudget string, gasFeeCharged string, storageFee string, storageRebate string, ) *L1EstimationResult`

NewL1EstimationResult instantiates a new L1EstimationResult object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewL1EstimationResultWithDefaults

`func NewL1EstimationResultWithDefaults() *L1EstimationResult`

NewL1EstimationResultWithDefaults instantiates a new L1EstimationResult object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetComputationFee

`func (o *L1EstimationResult) GetComputationFee() string`

GetComputationFee returns the ComputationFee field if non-nil, zero value otherwise.

### GetComputationFeeOk

`func (o *L1EstimationResult) GetComputationFeeOk() (*string, bool)`

GetComputationFeeOk returns a tuple with the ComputationFee field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComputationFee

`func (o *L1EstimationResult) SetComputationFee(v string)`

SetComputationFee sets ComputationFee field to given value.


### GetGasBudget

`func (o *L1EstimationResult) GetGasBudget() string`

GetGasBudget returns the GasBudget field if non-nil, zero value otherwise.

### GetGasBudgetOk

`func (o *L1EstimationResult) GetGasBudgetOk() (*string, bool)`

GetGasBudgetOk returns a tuple with the GasBudget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasBudget

`func (o *L1EstimationResult) SetGasBudget(v string)`

SetGasBudget sets GasBudget field to given value.


### GetGasFeeCharged

`func (o *L1EstimationResult) GetGasFeeCharged() string`

GetGasFeeCharged returns the GasFeeCharged field if non-nil, zero value otherwise.

### GetGasFeeChargedOk

`func (o *L1EstimationResult) GetGasFeeChargedOk() (*string, bool)`

GetGasFeeChargedOk returns a tuple with the GasFeeCharged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGasFeeCharged

`func (o *L1EstimationResult) SetGasFeeCharged(v string)`

SetGasFeeCharged sets GasFeeCharged field to given value.


### GetStorageFee

`func (o *L1EstimationResult) GetStorageFee() string`

GetStorageFee returns the StorageFee field if non-nil, zero value otherwise.

### GetStorageFeeOk

`func (o *L1EstimationResult) GetStorageFeeOk() (*string, bool)`

GetStorageFeeOk returns a tuple with the StorageFee field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageFee

`func (o *L1EstimationResult) SetStorageFee(v string)`

SetStorageFee sets StorageFee field to given value.


### GetStorageRebate

`func (o *L1EstimationResult) GetStorageRebate() string`

GetStorageRebate returns the StorageRebate field if non-nil, zero value otherwise.

### GetStorageRebateOk

`func (o *L1EstimationResult) GetStorageRebateOk() (*string, bool)`

GetStorageRebateOk returns a tuple with the StorageRebate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageRebate

`func (o *L1EstimationResult) SetStorageRebate(v string)`

SetStorageRebate sets StorageRebate field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


