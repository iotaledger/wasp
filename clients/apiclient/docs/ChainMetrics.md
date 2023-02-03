# ChainMetrics

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

### NewChainMetrics

`func NewChainMetrics(inAliasOutput AliasOutputMetricItem, inMilestone MilestoneMetricItem, inOnLedgerRequest OnLedgerRequestMetricItem, inOutput InOutputMetricItem, inStateOutput InStateOutputMetricItem, inTxInclusionState TxInclusionStateMsgMetricItem, outPublishGovernanceTransaction TransactionMetricItem, outPublisherStateTransaction PublisherStateTransactionItem, outPullLatestOutput InterfaceMetricItem, outPullOutputByID UTXOInputMetricItem, outPullTxInclusionState TransactionIDMetricItem, registeredChainIDs []string, ) *ChainMetrics`

NewChainMetrics instantiates a new ChainMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewChainMetricsWithDefaults

`func NewChainMetricsWithDefaults() *ChainMetrics`

NewChainMetricsWithDefaults instantiates a new ChainMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInAliasOutput

`func (o *ChainMetrics) GetInAliasOutput() AliasOutputMetricItem`

GetInAliasOutput returns the InAliasOutput field if non-nil, zero value otherwise.

### GetInAliasOutputOk

`func (o *ChainMetrics) GetInAliasOutputOk() (*AliasOutputMetricItem, bool)`

GetInAliasOutputOk returns a tuple with the InAliasOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInAliasOutput

`func (o *ChainMetrics) SetInAliasOutput(v AliasOutputMetricItem)`

SetInAliasOutput sets InAliasOutput field to given value.


### GetInMilestone

`func (o *ChainMetrics) GetInMilestone() MilestoneMetricItem`

GetInMilestone returns the InMilestone field if non-nil, zero value otherwise.

### GetInMilestoneOk

`func (o *ChainMetrics) GetInMilestoneOk() (*MilestoneMetricItem, bool)`

GetInMilestoneOk returns a tuple with the InMilestone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInMilestone

`func (o *ChainMetrics) SetInMilestone(v MilestoneMetricItem)`

SetInMilestone sets InMilestone field to given value.


### GetInOnLedgerRequest

`func (o *ChainMetrics) GetInOnLedgerRequest() OnLedgerRequestMetricItem`

GetInOnLedgerRequest returns the InOnLedgerRequest field if non-nil, zero value otherwise.

### GetInOnLedgerRequestOk

`func (o *ChainMetrics) GetInOnLedgerRequestOk() (*OnLedgerRequestMetricItem, bool)`

GetInOnLedgerRequestOk returns a tuple with the InOnLedgerRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInOnLedgerRequest

`func (o *ChainMetrics) SetInOnLedgerRequest(v OnLedgerRequestMetricItem)`

SetInOnLedgerRequest sets InOnLedgerRequest field to given value.


### GetInOutput

`func (o *ChainMetrics) GetInOutput() InOutputMetricItem`

GetInOutput returns the InOutput field if non-nil, zero value otherwise.

### GetInOutputOk

`func (o *ChainMetrics) GetInOutputOk() (*InOutputMetricItem, bool)`

GetInOutputOk returns a tuple with the InOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInOutput

`func (o *ChainMetrics) SetInOutput(v InOutputMetricItem)`

SetInOutput sets InOutput field to given value.


### GetInStateOutput

`func (o *ChainMetrics) GetInStateOutput() InStateOutputMetricItem`

GetInStateOutput returns the InStateOutput field if non-nil, zero value otherwise.

### GetInStateOutputOk

`func (o *ChainMetrics) GetInStateOutputOk() (*InStateOutputMetricItem, bool)`

GetInStateOutputOk returns a tuple with the InStateOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInStateOutput

`func (o *ChainMetrics) SetInStateOutput(v InStateOutputMetricItem)`

SetInStateOutput sets InStateOutput field to given value.


### GetInTxInclusionState

`func (o *ChainMetrics) GetInTxInclusionState() TxInclusionStateMsgMetricItem`

GetInTxInclusionState returns the InTxInclusionState field if non-nil, zero value otherwise.

### GetInTxInclusionStateOk

`func (o *ChainMetrics) GetInTxInclusionStateOk() (*TxInclusionStateMsgMetricItem, bool)`

GetInTxInclusionStateOk returns a tuple with the InTxInclusionState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInTxInclusionState

`func (o *ChainMetrics) SetInTxInclusionState(v TxInclusionStateMsgMetricItem)`

SetInTxInclusionState sets InTxInclusionState field to given value.


### GetOutPublishGovernanceTransaction

`func (o *ChainMetrics) GetOutPublishGovernanceTransaction() TransactionMetricItem`

GetOutPublishGovernanceTransaction returns the OutPublishGovernanceTransaction field if non-nil, zero value otherwise.

### GetOutPublishGovernanceTransactionOk

`func (o *ChainMetrics) GetOutPublishGovernanceTransactionOk() (*TransactionMetricItem, bool)`

GetOutPublishGovernanceTransactionOk returns a tuple with the OutPublishGovernanceTransaction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPublishGovernanceTransaction

`func (o *ChainMetrics) SetOutPublishGovernanceTransaction(v TransactionMetricItem)`

SetOutPublishGovernanceTransaction sets OutPublishGovernanceTransaction field to given value.


### GetOutPublisherStateTransaction

`func (o *ChainMetrics) GetOutPublisherStateTransaction() PublisherStateTransactionItem`

GetOutPublisherStateTransaction returns the OutPublisherStateTransaction field if non-nil, zero value otherwise.

### GetOutPublisherStateTransactionOk

`func (o *ChainMetrics) GetOutPublisherStateTransactionOk() (*PublisherStateTransactionItem, bool)`

GetOutPublisherStateTransactionOk returns a tuple with the OutPublisherStateTransaction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPublisherStateTransaction

`func (o *ChainMetrics) SetOutPublisherStateTransaction(v PublisherStateTransactionItem)`

SetOutPublisherStateTransaction sets OutPublisherStateTransaction field to given value.


### GetOutPullLatestOutput

`func (o *ChainMetrics) GetOutPullLatestOutput() InterfaceMetricItem`

GetOutPullLatestOutput returns the OutPullLatestOutput field if non-nil, zero value otherwise.

### GetOutPullLatestOutputOk

`func (o *ChainMetrics) GetOutPullLatestOutputOk() (*InterfaceMetricItem, bool)`

GetOutPullLatestOutputOk returns a tuple with the OutPullLatestOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullLatestOutput

`func (o *ChainMetrics) SetOutPullLatestOutput(v InterfaceMetricItem)`

SetOutPullLatestOutput sets OutPullLatestOutput field to given value.


### GetOutPullOutputByID

`func (o *ChainMetrics) GetOutPullOutputByID() UTXOInputMetricItem`

GetOutPullOutputByID returns the OutPullOutputByID field if non-nil, zero value otherwise.

### GetOutPullOutputByIDOk

`func (o *ChainMetrics) GetOutPullOutputByIDOk() (*UTXOInputMetricItem, bool)`

GetOutPullOutputByIDOk returns a tuple with the OutPullOutputByID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullOutputByID

`func (o *ChainMetrics) SetOutPullOutputByID(v UTXOInputMetricItem)`

SetOutPullOutputByID sets OutPullOutputByID field to given value.


### GetOutPullTxInclusionState

`func (o *ChainMetrics) GetOutPullTxInclusionState() TransactionIDMetricItem`

GetOutPullTxInclusionState returns the OutPullTxInclusionState field if non-nil, zero value otherwise.

### GetOutPullTxInclusionStateOk

`func (o *ChainMetrics) GetOutPullTxInclusionStateOk() (*TransactionIDMetricItem, bool)`

GetOutPullTxInclusionStateOk returns a tuple with the OutPullTxInclusionState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutPullTxInclusionState

`func (o *ChainMetrics) SetOutPullTxInclusionState(v TransactionIDMetricItem)`

SetOutPullTxInclusionState sets OutPullTxInclusionState field to given value.


### GetRegisteredChainIDs

`func (o *ChainMetrics) GetRegisteredChainIDs() []string`

GetRegisteredChainIDs returns the RegisteredChainIDs field if non-nil, zero value otherwise.

### GetRegisteredChainIDsOk

`func (o *ChainMetrics) GetRegisteredChainIDsOk() (*[]string, bool)`

GetRegisteredChainIDsOk returns a tuple with the RegisteredChainIDs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegisteredChainIDs

`func (o *ChainMetrics) SetRegisteredChainIDs(v []string)`

SetRegisteredChainIDs sets RegisteredChainIDs field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


