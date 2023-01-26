# ChainMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InAliasOutput** | Pointer to [**AliasOutputMetricItem**](AliasOutputMetricItem.md) |  | [optional] 
**InOnLedgerRequest** | Pointer to [**OnLedgerRequestMetricItem**](OnLedgerRequestMetricItem.md) |  | [optional] 
**InOutput** | Pointer to [**InOutputMetricItem**](InOutputMetricItem.md) |  | [optional] 
**InStateOutput** | Pointer to [**InStateOutputMetricItem**](InStateOutputMetricItem.md) |  | [optional] 
**InTxInclusionState** | Pointer to [**TxInclusionStateMsgMetricItem**](TxInclusionStateMsgMetricItem.md) |  | [optional] 
**OutPublishGovernanceTransaction** | Pointer to [**TransactionMetricItem**](TransactionMetricItem.md) |  | [optional] 
**OutPullLatestOutput** | Pointer to [**InterfaceMetricItem**](InterfaceMetricItem.md) |  | [optional] 
**OutPullOutputByID** | Pointer to [**UTXOInputMetricItem**](UTXOInputMetricItem.md) |  | [optional] 
**OutPullTxInclusionState** | Pointer to [**TransactionIDMetricItem**](TransactionIDMetricItem.md) |  | [optional] 

## Methods

### NewChainMetrics

`func NewChainMetrics() *ChainMetrics`

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

### HasInAliasOutput

`func (o *ChainMetrics) HasInAliasOutput() bool`

HasInAliasOutput returns a boolean if a field has been set.

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

### HasInOnLedgerRequest

`func (o *ChainMetrics) HasInOnLedgerRequest() bool`

HasInOnLedgerRequest returns a boolean if a field has been set.

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

### HasInOutput

`func (o *ChainMetrics) HasInOutput() bool`

HasInOutput returns a boolean if a field has been set.

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

### HasInStateOutput

`func (o *ChainMetrics) HasInStateOutput() bool`

HasInStateOutput returns a boolean if a field has been set.

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

### HasInTxInclusionState

`func (o *ChainMetrics) HasInTxInclusionState() bool`

HasInTxInclusionState returns a boolean if a field has been set.

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

### HasOutPublishGovernanceTransaction

`func (o *ChainMetrics) HasOutPublishGovernanceTransaction() bool`

HasOutPublishGovernanceTransaction returns a boolean if a field has been set.

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

### HasOutPullLatestOutput

`func (o *ChainMetrics) HasOutPullLatestOutput() bool`

HasOutPullLatestOutput returns a boolean if a field has been set.

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

### HasOutPullOutputByID

`func (o *ChainMetrics) HasOutPullOutputByID() bool`

HasOutPullOutputByID returns a boolean if a field has been set.

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

### HasOutPullTxInclusionState

`func (o *ChainMetrics) HasOutPullTxInclusionState() bool`

HasOutPullTxInclusionState returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


