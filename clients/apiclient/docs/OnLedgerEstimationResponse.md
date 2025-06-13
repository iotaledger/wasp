# OnLedgerEstimationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**L1** | [**L1EstimationResult**](L1EstimationResult.md) |  | 
**L2** | [**ReceiptResponse**](ReceiptResponse.md) |  | 

## Methods

### NewOnLedgerEstimationResponse

`func NewOnLedgerEstimationResponse(l1 L1EstimationResult, l2 ReceiptResponse, ) *OnLedgerEstimationResponse`

NewOnLedgerEstimationResponse instantiates a new OnLedgerEstimationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOnLedgerEstimationResponseWithDefaults

`func NewOnLedgerEstimationResponseWithDefaults() *OnLedgerEstimationResponse`

NewOnLedgerEstimationResponseWithDefaults instantiates a new OnLedgerEstimationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetL1

`func (o *OnLedgerEstimationResponse) GetL1() L1EstimationResult`

GetL1 returns the L1 field if non-nil, zero value otherwise.

### GetL1Ok

`func (o *OnLedgerEstimationResponse) GetL1Ok() (*L1EstimationResult, bool)`

GetL1Ok returns a tuple with the L1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetL1

`func (o *OnLedgerEstimationResponse) SetL1(v L1EstimationResult)`

SetL1 sets L1 field to given value.


### GetL2

`func (o *OnLedgerEstimationResponse) GetL2() ReceiptResponse`

GetL2 returns the L2 field if non-nil, zero value otherwise.

### GetL2Ok

`func (o *OnLedgerEstimationResponse) GetL2Ok() (*ReceiptResponse, bool)`

GetL2Ok returns a tuple with the L2 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetL2

`func (o *OnLedgerEstimationResponse) SetL2(v ReceiptResponse)`

SetL2 sets L2 field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


