# ChainMessageMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InAnchor** | [**AnchorMetricItem**](AnchorMetricItem.md) |  | 
**InOnLedgerRequest** | [**OnLedgerRequestMetricItem**](OnLedgerRequestMetricItem.md) |  | 
**OutPublisherStateTransaction** | [**PublisherStateTransactionItem**](PublisherStateTransactionItem.md) |  | 

## Methods

### NewChainMessageMetrics

`func NewChainMessageMetrics(inAnchor AnchorMetricItem, inOnLedgerRequest OnLedgerRequestMetricItem, outPublisherStateTransaction PublisherStateTransactionItem, ) *ChainMessageMetrics`

NewChainMessageMetrics instantiates a new ChainMessageMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewChainMessageMetricsWithDefaults

`func NewChainMessageMetricsWithDefaults() *ChainMessageMetrics`

NewChainMessageMetricsWithDefaults instantiates a new ChainMessageMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInAnchor

`func (o *ChainMessageMetrics) GetInAnchor() AnchorMetricItem`

GetInAnchor returns the InAnchor field if non-nil, zero value otherwise.

### GetInAnchorOk

`func (o *ChainMessageMetrics) GetInAnchorOk() (*AnchorMetricItem, bool)`

GetInAnchorOk returns a tuple with the InAnchor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInAnchor

`func (o *ChainMessageMetrics) SetInAnchor(v AnchorMetricItem)`

SetInAnchor sets InAnchor field to given value.


### GetInOnLedgerRequest

`func (o *ChainMessageMetrics) GetInOnLedgerRequest() OnLedgerRequestMetricItem`

GetInOnLedgerRequest returns the InOnLedgerRequest field if non-nil, zero value otherwise.

### GetInOnLedgerRequestOk

`func (o *ChainMessageMetrics) GetInOnLedgerRequestOk() (*OnLedgerRequestMetricItem, bool)`

GetInOnLedgerRequestOk returns a tuple with the InOnLedgerRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInOnLedgerRequest

`func (o *ChainMessageMetrics) SetInOnLedgerRequest(v OnLedgerRequestMetricItem)`

SetInOnLedgerRequest sets InOnLedgerRequest field to given value.


### GetOutPublisherStateTransaction

`func (o *ChainMessageMetrics) GetOutPublisherStateTransaction() PublisherStateTransactionItem`

GetOutPublisherStateTransaction returns the OutPublisherStateTransaction field if non-nil, zero value otherwise.

### GetOutPublisherStateTransactionOk

`func (o *ChainMessageMetrics) GetOutPublisherStateTransactionOk() (*PublisherStateTransactionItem, bool)`

GetOutPublisherStateTransactionOk returns a tuple with the OutPublisherStateTransaction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPublisherStateTransaction

`func (o *ChainMessageMetrics) SetOutPublisherStateTransaction(v PublisherStateTransactionItem)`

SetOutPublisherStateTransaction sets OutPublisherStateTransaction field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


