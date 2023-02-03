# WaitRequestProcessedParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Timeout** | Pointer to **int64** | Timeout in nanoseconds | [optional] 

## Methods

### NewWaitRequestProcessedParams

`func NewWaitRequestProcessedParams() *WaitRequestProcessedParams`

NewWaitRequestProcessedParams instantiates a new WaitRequestProcessedParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWaitRequestProcessedParamsWithDefaults

`func NewWaitRequestProcessedParamsWithDefaults() *WaitRequestProcessedParams`

NewWaitRequestProcessedParamsWithDefaults instantiates a new WaitRequestProcessedParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTimeout

`func (o *WaitRequestProcessedParams) GetTimeout() int64`

GetTimeout returns the Timeout field if non-nil, zero value otherwise.

### GetTimeoutOk

`func (o *WaitRequestProcessedParams) GetTimeoutOk() (*int64, bool)`

GetTimeoutOk returns a tuple with the Timeout field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeout

`func (o *WaitRequestProcessedParams) SetTimeout(v int64)`

SetTimeout sets Timeout field to given value.

### HasTimeout

`func (o *WaitRequestProcessedParams) HasTimeout() bool`

HasTimeout returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


