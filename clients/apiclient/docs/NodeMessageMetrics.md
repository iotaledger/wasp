# NodeMessageMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InAnchor** | [**AnchorMetricItem**](AnchorMetricItem.md) |  | 
**InOnLedgerRequest** | [**OnLedgerRequestMetricItem**](OnLedgerRequestMetricItem.md) |  | 
**OutPublisherStateTransaction** | [**PublisherStateTransactionItem**](PublisherStateTransactionItem.md) |  | 
**RegisteredChainIDs** | **[]string** |  | 

## Methods

### NewNodeMessageMetrics

`func NewNodeMessageMetrics(inAnchor AnchorMetricItem, inOnLedgerRequest OnLedgerRequestMetricItem, outPublisherStateTransaction PublisherStateTransactionItem, registeredChainIDs []string, ) *NodeMessageMetrics`

NewNodeMessageMetrics instantiates a new NodeMessageMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeMessageMetricsWithDefaults

`func NewNodeMessageMetricsWithDefaults() *NodeMessageMetrics`

NewNodeMessageMetricsWithDefaults instantiates a new NodeMessageMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInAnchor

`func (o *NodeMessageMetrics) GetInAnchor() AnchorMetricItem`

GetInAnchor returns the InAnchor field if non-nil, zero value otherwise.

### GetInAnchorOk

`func (o *NodeMessageMetrics) GetInAnchorOk() (*AnchorMetricItem, bool)`

GetInAnchorOk returns a tuple with the InAnchor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInAnchor

`func (o *NodeMessageMetrics) SetInAnchor(v AnchorMetricItem)`

SetInAnchor sets InAnchor field to given value.


### GetInOnLedgerRequest

`func (o *NodeMessageMetrics) GetInOnLedgerRequest() OnLedgerRequestMetricItem`

GetInOnLedgerRequest returns the InOnLedgerRequest field if non-nil, zero value otherwise.

### GetInOnLedgerRequestOk

`func (o *NodeMessageMetrics) GetInOnLedgerRequestOk() (*OnLedgerRequestMetricItem, bool)`

GetInOnLedgerRequestOk returns a tuple with the InOnLedgerRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInOnLedgerRequest

`func (o *NodeMessageMetrics) SetInOnLedgerRequest(v OnLedgerRequestMetricItem)`

SetInOnLedgerRequest sets InOnLedgerRequest field to given value.


### GetOutPublisherStateTransaction

`func (o *NodeMessageMetrics) GetOutPublisherStateTransaction() PublisherStateTransactionItem`

GetOutPublisherStateTransaction returns the OutPublisherStateTransaction field if non-nil, zero value otherwise.

### GetOutPublisherStateTransactionOk

`func (o *NodeMessageMetrics) GetOutPublisherStateTransactionOk() (*PublisherStateTransactionItem, bool)`

GetOutPublisherStateTransactionOk returns a tuple with the OutPublisherStateTransaction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPublisherStateTransaction

`func (o *NodeMessageMetrics) SetOutPublisherStateTransaction(v PublisherStateTransactionItem)`

SetOutPublisherStateTransaction sets OutPublisherStateTransaction field to given value.


### GetRegisteredChainIDs

`func (o *NodeMessageMetrics) GetRegisteredChainIDs() []string`

GetRegisteredChainIDs returns the RegisteredChainIDs field if non-nil, zero value otherwise.

### GetRegisteredChainIDsOk

`func (o *NodeMessageMetrics) GetRegisteredChainIDsOk() (*[]string, bool)`

GetRegisteredChainIDsOk returns a tuple with the RegisteredChainIDs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegisteredChainIDs

`func (o *NodeMessageMetrics) SetRegisteredChainIDs(v []string)`

SetRegisteredChainIDs sets RegisteredChainIDs field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


