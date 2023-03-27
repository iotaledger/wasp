# ChainMessageMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InAliasOutput** | [**AliasOutputMetricItem**](AliasOutputMetricItem.md) |  | 
**InOnLedgerRequest** | [**OnLedgerRequestMetricItem**](OnLedgerRequestMetricItem.md) |  | 
**InOutput** | [**InOutputMetricItem**](InOutputMetricItem.md) |  | 
**InStateOutput** | [**InStateOutputMetricItem**](InStateOutputMetricItem.md) |  | 
**InTxInclusionState** | [**TxInclusionStateMsgMetricItem**](TxInclusionStateMsgMetricItem.md) |  | 
**OutPublishGovernanceTransaction** | [**TransactionMetricItem**](TransactionMetricItem.md) |  | 
**OutPublisherStateTransaction** | [**PublisherStateTransactionItem**](PublisherStateTransactionItem.md) |  | 
**OutPullLatestOutput** | [**InterfaceMetricItem**](InterfaceMetricItem.md) |  | 
**OutPullOutputByID** | [**UTXOInputMetricItem**](UTXOInputMetricItem.md) |  | 
**OutPullTxInclusionState** | [**TransactionIDMetricItem**](TransactionIDMetricItem.md) |  | 

## Methods

### NewChainMessageMetrics

`func NewChainMessageMetrics(inAliasOutput AliasOutputMetricItem, inOnLedgerRequest OnLedgerRequestMetricItem, inOutput InOutputMetricItem, inStateOutput InStateOutputMetricItem, inTxInclusionState TxInclusionStateMsgMetricItem, outPublishGovernanceTransaction TransactionMetricItem, outPublisherStateTransaction PublisherStateTransactionItem, outPullLatestOutput InterfaceMetricItem, outPullOutputByID UTXOInputMetricItem, outPullTxInclusionState TransactionIDMetricItem, ) *ChainMessageMetrics`

NewChainMessageMetrics instantiates a new ChainMessageMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewChainMessageMetricsWithDefaults

`func NewChainMessageMetricsWithDefaults() *ChainMessageMetrics`

NewChainMessageMetricsWithDefaults instantiates a new ChainMessageMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInAliasOutput

`func (o *ChainMessageMetrics) GetInAliasOutput() AliasOutputMetricItem`

GetInAliasOutput returns the InAliasOutput field if non-nil, zero value otherwise.

### GetInAliasOutputOk

`func (o *ChainMessageMetrics) GetInAliasOutputOk() (*AliasOutputMetricItem, bool)`

GetInAliasOutputOk returns a tuple with the InAliasOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInAliasOutput

`func (o *ChainMessageMetrics) SetInAliasOutput(v AliasOutputMetricItem)`

SetInAliasOutput sets InAliasOutput field to given value.


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


### GetInOutput

`func (o *ChainMessageMetrics) GetInOutput() InOutputMetricItem`

GetInOutput returns the InOutput field if non-nil, zero value otherwise.

### GetInOutputOk

`func (o *ChainMessageMetrics) GetInOutputOk() (*InOutputMetricItem, bool)`

GetInOutputOk returns a tuple with the InOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInOutput

`func (o *ChainMessageMetrics) SetInOutput(v InOutputMetricItem)`

SetInOutput sets InOutput field to given value.


### GetInStateOutput

`func (o *ChainMessageMetrics) GetInStateOutput() InStateOutputMetricItem`

GetInStateOutput returns the InStateOutput field if non-nil, zero value otherwise.

### GetInStateOutputOk

`func (o *ChainMessageMetrics) GetInStateOutputOk() (*InStateOutputMetricItem, bool)`

GetInStateOutputOk returns a tuple with the InStateOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInStateOutput

`func (o *ChainMessageMetrics) SetInStateOutput(v InStateOutputMetricItem)`

SetInStateOutput sets InStateOutput field to given value.


### GetInTxInclusionState

`func (o *ChainMessageMetrics) GetInTxInclusionState() TxInclusionStateMsgMetricItem`

GetInTxInclusionState returns the InTxInclusionState field if non-nil, zero value otherwise.

### GetInTxInclusionStateOk

`func (o *ChainMessageMetrics) GetInTxInclusionStateOk() (*TxInclusionStateMsgMetricItem, bool)`

GetInTxInclusionStateOk returns a tuple with the InTxInclusionState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInTxInclusionState

`func (o *ChainMessageMetrics) SetInTxInclusionState(v TxInclusionStateMsgMetricItem)`

SetInTxInclusionState sets InTxInclusionState field to given value.


### GetOutPublishGovernanceTransaction

`func (o *ChainMessageMetrics) GetOutPublishGovernanceTransaction() TransactionMetricItem`

GetOutPublishGovernanceTransaction returns the OutPublishGovernanceTransaction field if non-nil, zero value otherwise.

### GetOutPublishGovernanceTransactionOk

`func (o *ChainMessageMetrics) GetOutPublishGovernanceTransactionOk() (*TransactionMetricItem, bool)`

GetOutPublishGovernanceTransactionOk returns a tuple with the OutPublishGovernanceTransaction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPublishGovernanceTransaction

`func (o *ChainMessageMetrics) SetOutPublishGovernanceTransaction(v TransactionMetricItem)`

SetOutPublishGovernanceTransaction sets OutPublishGovernanceTransaction field to given value.


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


### GetOutPullLatestOutput

`func (o *ChainMessageMetrics) GetOutPullLatestOutput() InterfaceMetricItem`

GetOutPullLatestOutput returns the OutPullLatestOutput field if non-nil, zero value otherwise.

### GetOutPullLatestOutputOk

`func (o *ChainMessageMetrics) GetOutPullLatestOutputOk() (*InterfaceMetricItem, bool)`

GetOutPullLatestOutputOk returns a tuple with the OutPullLatestOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullLatestOutput

`func (o *ChainMessageMetrics) SetOutPullLatestOutput(v InterfaceMetricItem)`

SetOutPullLatestOutput sets OutPullLatestOutput field to given value.


### GetOutPullOutputByID

`func (o *ChainMessageMetrics) GetOutPullOutputByID() UTXOInputMetricItem`

GetOutPullOutputByID returns the OutPullOutputByID field if non-nil, zero value otherwise.

### GetOutPullOutputByIDOk

`func (o *ChainMessageMetrics) GetOutPullOutputByIDOk() (*UTXOInputMetricItem, bool)`

GetOutPullOutputByIDOk returns a tuple with the OutPullOutputByID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullOutputByID

`func (o *ChainMessageMetrics) SetOutPullOutputByID(v UTXOInputMetricItem)`

SetOutPullOutputByID sets OutPullOutputByID field to given value.


### GetOutPullTxInclusionState

`func (o *ChainMessageMetrics) GetOutPullTxInclusionState() TransactionIDMetricItem`

GetOutPullTxInclusionState returns the OutPullTxInclusionState field if non-nil, zero value otherwise.

### GetOutPullTxInclusionStateOk

`func (o *ChainMessageMetrics) GetOutPullTxInclusionStateOk() (*TransactionIDMetricItem, bool)`

GetOutPullTxInclusionStateOk returns a tuple with the OutPullTxInclusionState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullTxInclusionState

`func (o *ChainMessageMetrics) SetOutPullTxInclusionState(v TransactionIDMetricItem)`

SetOutPullTxInclusionState sets OutPullTxInclusionState field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


