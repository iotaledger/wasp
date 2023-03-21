# NodeMessageMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InAliasOutput** | [**AliasOutputMetricItem**](AliasOutputMetricItem.md) |  | 
**InMilestone** | [**MilestoneMetricItem**](MilestoneMetricItem.md) |  | 
**InOnLedgerRequest** | [**OnLedgerRequestMetricItem**](OnLedgerRequestMetricItem.md) |  | 
**InOutput** | [**InOutputMetricItem**](InOutputMetricItem.md) |  | 
**InStateOutput** | [**InStateOutputMetricItem**](InStateOutputMetricItem.md) |  | 
**InTxInclusionState** | [**TxInclusionStateMsgMetricItem**](TxInclusionStateMsgMetricItem.md) |  | 
**OutPublishGovernanceTransaction** | [**TransactionMetricItem**](TransactionMetricItem.md) |  | 
**OutPublisherStateTransaction** | [**PublisherStateTransactionItem**](PublisherStateTransactionItem.md) |  | 
**OutPullLatestOutput** | [**InterfaceMetricItem**](InterfaceMetricItem.md) |  | 
**OutPullOutputByID** | [**UTXOInputMetricItem**](UTXOInputMetricItem.md) |  | 
**OutPullTxInclusionState** | [**TransactionIDMetricItem**](TransactionIDMetricItem.md) |  | 
**RegisteredChainIDs** | **[]string** |  | 

## Methods

### NewNodeMessageMetrics

`func NewNodeMessageMetrics(inAliasOutput AliasOutputMetricItem, inMilestone MilestoneMetricItem, inOnLedgerRequest OnLedgerRequestMetricItem, inOutput InOutputMetricItem, inStateOutput InStateOutputMetricItem, inTxInclusionState TxInclusionStateMsgMetricItem, outPublishGovernanceTransaction TransactionMetricItem, outPublisherStateTransaction PublisherStateTransactionItem, outPullLatestOutput InterfaceMetricItem, outPullOutputByID UTXOInputMetricItem, outPullTxInclusionState TransactionIDMetricItem, registeredChainIDs []string, ) *NodeMessageMetrics`

NewNodeMessageMetrics instantiates a new NodeMessageMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeMessageMetricsWithDefaults

`func NewNodeMessageMetricsWithDefaults() *NodeMessageMetrics`

NewNodeMessageMetricsWithDefaults instantiates a new NodeMessageMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInAliasOutput

`func (o *NodeMessageMetrics) GetInAliasOutput() AliasOutputMetricItem`

GetInAliasOutput returns the InAliasOutput field if non-nil, zero value otherwise.

### GetInAliasOutputOk

`func (o *NodeMessageMetrics) GetInAliasOutputOk() (*AliasOutputMetricItem, bool)`

GetInAliasOutputOk returns a tuple with the InAliasOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInAliasOutput

`func (o *NodeMessageMetrics) SetInAliasOutput(v AliasOutputMetricItem)`

SetInAliasOutput sets InAliasOutput field to given value.


### GetInMilestone

`func (o *NodeMessageMetrics) GetInMilestone() MilestoneMetricItem`

GetInMilestone returns the InMilestone field if non-nil, zero value otherwise.

### GetInMilestoneOk

`func (o *NodeMessageMetrics) GetInMilestoneOk() (*MilestoneMetricItem, bool)`

GetInMilestoneOk returns a tuple with the InMilestone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInMilestone

`func (o *NodeMessageMetrics) SetInMilestone(v MilestoneMetricItem)`

SetInMilestone sets InMilestone field to given value.


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


### GetInOutput

`func (o *NodeMessageMetrics) GetInOutput() InOutputMetricItem`

GetInOutput returns the InOutput field if non-nil, zero value otherwise.

### GetInOutputOk

`func (o *NodeMessageMetrics) GetInOutputOk() (*InOutputMetricItem, bool)`

GetInOutputOk returns a tuple with the InOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInOutput

`func (o *NodeMessageMetrics) SetInOutput(v InOutputMetricItem)`

SetInOutput sets InOutput field to given value.


### GetInStateOutput

`func (o *NodeMessageMetrics) GetInStateOutput() InStateOutputMetricItem`

GetInStateOutput returns the InStateOutput field if non-nil, zero value otherwise.

### GetInStateOutputOk

`func (o *NodeMessageMetrics) GetInStateOutputOk() (*InStateOutputMetricItem, bool)`

GetInStateOutputOk returns a tuple with the InStateOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInStateOutput

`func (o *NodeMessageMetrics) SetInStateOutput(v InStateOutputMetricItem)`

SetInStateOutput sets InStateOutput field to given value.


### GetInTxInclusionState

`func (o *NodeMessageMetrics) GetInTxInclusionState() TxInclusionStateMsgMetricItem`

GetInTxInclusionState returns the InTxInclusionState field if non-nil, zero value otherwise.

### GetInTxInclusionStateOk

`func (o *NodeMessageMetrics) GetInTxInclusionStateOk() (*TxInclusionStateMsgMetricItem, bool)`

GetInTxInclusionStateOk returns a tuple with the InTxInclusionState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInTxInclusionState

`func (o *NodeMessageMetrics) SetInTxInclusionState(v TxInclusionStateMsgMetricItem)`

SetInTxInclusionState sets InTxInclusionState field to given value.


### GetOutPublishGovernanceTransaction

`func (o *NodeMessageMetrics) GetOutPublishGovernanceTransaction() TransactionMetricItem`

GetOutPublishGovernanceTransaction returns the OutPublishGovernanceTransaction field if non-nil, zero value otherwise.

### GetOutPublishGovernanceTransactionOk

`func (o *NodeMessageMetrics) GetOutPublishGovernanceTransactionOk() (*TransactionMetricItem, bool)`

GetOutPublishGovernanceTransactionOk returns a tuple with the OutPublishGovernanceTransaction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPublishGovernanceTransaction

`func (o *NodeMessageMetrics) SetOutPublishGovernanceTransaction(v TransactionMetricItem)`

SetOutPublishGovernanceTransaction sets OutPublishGovernanceTransaction field to given value.


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


### GetOutPullLatestOutput

`func (o *NodeMessageMetrics) GetOutPullLatestOutput() InterfaceMetricItem`

GetOutPullLatestOutput returns the OutPullLatestOutput field if non-nil, zero value otherwise.

### GetOutPullLatestOutputOk

`func (o *NodeMessageMetrics) GetOutPullLatestOutputOk() (*InterfaceMetricItem, bool)`

GetOutPullLatestOutputOk returns a tuple with the OutPullLatestOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullLatestOutput

`func (o *NodeMessageMetrics) SetOutPullLatestOutput(v InterfaceMetricItem)`

SetOutPullLatestOutput sets OutPullLatestOutput field to given value.


### GetOutPullOutputByID

`func (o *NodeMessageMetrics) GetOutPullOutputByID() UTXOInputMetricItem`

GetOutPullOutputByID returns the OutPullOutputByID field if non-nil, zero value otherwise.

### GetOutPullOutputByIDOk

`func (o *NodeMessageMetrics) GetOutPullOutputByIDOk() (*UTXOInputMetricItem, bool)`

GetOutPullOutputByIDOk returns a tuple with the OutPullOutputByID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullOutputByID

`func (o *NodeMessageMetrics) SetOutPullOutputByID(v UTXOInputMetricItem)`

SetOutPullOutputByID sets OutPullOutputByID field to given value.


### GetOutPullTxInclusionState

`func (o *NodeMessageMetrics) GetOutPullTxInclusionState() TransactionIDMetricItem`

GetOutPullTxInclusionState returns the OutPullTxInclusionState field if non-nil, zero value otherwise.

### GetOutPullTxInclusionStateOk

`func (o *NodeMessageMetrics) GetOutPullTxInclusionStateOk() (*TransactionIDMetricItem, bool)`

GetOutPullTxInclusionStateOk returns a tuple with the OutPullTxInclusionState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullTxInclusionState

`func (o *NodeMessageMetrics) SetOutPullTxInclusionState(v TransactionIDMetricItem)`

SetOutPullTxInclusionState sets OutPullTxInclusionState field to given value.


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


