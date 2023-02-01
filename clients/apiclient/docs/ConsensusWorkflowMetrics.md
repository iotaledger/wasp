# ConsensusWorkflowMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CurrentStateIndex** | Pointer to **uint32** | Shows current state index of the consensus | [optional] 
**FlagBatchProposalSent** | **bool** | Shows if batch proposal is sent out in current consensus iteration | 
**FlagConsensusBatchKnown** | **bool** | Shows if consensus on batch is reached and known in current consensus iteration | 
**FlagInProgress** | **bool** | Shows if consensus algorithm is still not completed in current consensus iteration | 
**FlagStateReceived** | **bool** | Shows if state output is received in current consensus iteration | 
**FlagTransactionFinalized** | **bool** | Shows if consensus on transaction is reached in current consensus iteration | 
**FlagTransactionPosted** | **bool** | Shows if transaction is posted to L1 in current consensus iteration | 
**FlagTransactionSeen** | **bool** | Shows if L1 reported that it has seen the transaction of current consensus iteration | 
**FlagVMResultSigned** | **bool** | Shows if virtual machine has returned its results in current consensus iteration | 
**FlagVMStarted** | **bool** | Shows if virtual machine is started in current consensus iteration | 
**TimeBatchProposalSent** | **time.Time** | Shows when batch proposal was last sent out in current consensus iteration | 
**TimeCompleted** | **time.Time** | Shows when algorithm was last completed in current consensus iteration | 
**TimeConsensusBatchKnown** | **time.Time** | Shows when ACS results of consensus on batch was last received in current consensus iteration | 
**TimeTransactionFinalized** | **time.Time** | Shows when algorithm last noted that all the data for consensus on transaction had been received in current consensus iteration | 
**TimeTransactionPosted** | **time.Time** | Shows when transaction was last posted to L1 in current consensus iteration | 
**TimeTransactionSeen** | **time.Time** | Shows when algorithm last noted that transaction had been seen by L1 in current consensus iteration | 
**TimeVMResultSigned** | **time.Time** | Shows when virtual machine results were last received and signed in current consensus iteration | 
**TimeVMStarted** | **time.Time** | Shows when virtual machine was last started in current consensus iteration | 

## Methods

### NewConsensusWorkflowMetrics

`func NewConsensusWorkflowMetrics(flagBatchProposalSent bool, flagConsensusBatchKnown bool, flagInProgress bool, flagStateReceived bool, flagTransactionFinalized bool, flagTransactionPosted bool, flagTransactionSeen bool, flagVMResultSigned bool, flagVMStarted bool, timeBatchProposalSent time.Time, timeCompleted time.Time, timeConsensusBatchKnown time.Time, timeTransactionFinalized time.Time, timeTransactionPosted time.Time, timeTransactionSeen time.Time, timeVMResultSigned time.Time, timeVMStarted time.Time, ) *ConsensusWorkflowMetrics`

NewConsensusWorkflowMetrics instantiates a new ConsensusWorkflowMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConsensusWorkflowMetricsWithDefaults

`func NewConsensusWorkflowMetricsWithDefaults() *ConsensusWorkflowMetrics`

NewConsensusWorkflowMetricsWithDefaults instantiates a new ConsensusWorkflowMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCurrentStateIndex

`func (o *ConsensusWorkflowMetrics) GetCurrentStateIndex() uint32`

GetCurrentStateIndex returns the CurrentStateIndex field if non-nil, zero value otherwise.

### GetCurrentStateIndexOk

`func (o *ConsensusWorkflowMetrics) GetCurrentStateIndexOk() (*uint32, bool)`

GetCurrentStateIndexOk returns a tuple with the CurrentStateIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentStateIndex

`func (o *ConsensusWorkflowMetrics) SetCurrentStateIndex(v uint32)`

SetCurrentStateIndex sets CurrentStateIndex field to given value.

### HasCurrentStateIndex

`func (o *ConsensusWorkflowMetrics) HasCurrentStateIndex() bool`

HasCurrentStateIndex returns a boolean if a field has been set.

### GetFlagBatchProposalSent

`func (o *ConsensusWorkflowMetrics) GetFlagBatchProposalSent() bool`

GetFlagBatchProposalSent returns the FlagBatchProposalSent field if non-nil, zero value otherwise.

### GetFlagBatchProposalSentOk

`func (o *ConsensusWorkflowMetrics) GetFlagBatchProposalSentOk() (*bool, bool)`

GetFlagBatchProposalSentOk returns a tuple with the FlagBatchProposalSent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagBatchProposalSent

`func (o *ConsensusWorkflowMetrics) SetFlagBatchProposalSent(v bool)`

SetFlagBatchProposalSent sets FlagBatchProposalSent field to given value.


### GetFlagConsensusBatchKnown

`func (o *ConsensusWorkflowMetrics) GetFlagConsensusBatchKnown() bool`

GetFlagConsensusBatchKnown returns the FlagConsensusBatchKnown field if non-nil, zero value otherwise.

### GetFlagConsensusBatchKnownOk

`func (o *ConsensusWorkflowMetrics) GetFlagConsensusBatchKnownOk() (*bool, bool)`

GetFlagConsensusBatchKnownOk returns a tuple with the FlagConsensusBatchKnown field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagConsensusBatchKnown

`func (o *ConsensusWorkflowMetrics) SetFlagConsensusBatchKnown(v bool)`

SetFlagConsensusBatchKnown sets FlagConsensusBatchKnown field to given value.


### GetFlagInProgress

`func (o *ConsensusWorkflowMetrics) GetFlagInProgress() bool`

GetFlagInProgress returns the FlagInProgress field if non-nil, zero value otherwise.

### GetFlagInProgressOk

`func (o *ConsensusWorkflowMetrics) GetFlagInProgressOk() (*bool, bool)`

GetFlagInProgressOk returns a tuple with the FlagInProgress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagInProgress

`func (o *ConsensusWorkflowMetrics) SetFlagInProgress(v bool)`

SetFlagInProgress sets FlagInProgress field to given value.


### GetFlagStateReceived

`func (o *ConsensusWorkflowMetrics) GetFlagStateReceived() bool`

GetFlagStateReceived returns the FlagStateReceived field if non-nil, zero value otherwise.

### GetFlagStateReceivedOk

`func (o *ConsensusWorkflowMetrics) GetFlagStateReceivedOk() (*bool, bool)`

GetFlagStateReceivedOk returns a tuple with the FlagStateReceived field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagStateReceived

`func (o *ConsensusWorkflowMetrics) SetFlagStateReceived(v bool)`

SetFlagStateReceived sets FlagStateReceived field to given value.


### GetFlagTransactionFinalized

`func (o *ConsensusWorkflowMetrics) GetFlagTransactionFinalized() bool`

GetFlagTransactionFinalized returns the FlagTransactionFinalized field if non-nil, zero value otherwise.

### GetFlagTransactionFinalizedOk

`func (o *ConsensusWorkflowMetrics) GetFlagTransactionFinalizedOk() (*bool, bool)`

GetFlagTransactionFinalizedOk returns a tuple with the FlagTransactionFinalized field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagTransactionFinalized

`func (o *ConsensusWorkflowMetrics) SetFlagTransactionFinalized(v bool)`

SetFlagTransactionFinalized sets FlagTransactionFinalized field to given value.


### GetFlagTransactionPosted

`func (o *ConsensusWorkflowMetrics) GetFlagTransactionPosted() bool`

GetFlagTransactionPosted returns the FlagTransactionPosted field if non-nil, zero value otherwise.

### GetFlagTransactionPostedOk

`func (o *ConsensusWorkflowMetrics) GetFlagTransactionPostedOk() (*bool, bool)`

GetFlagTransactionPostedOk returns a tuple with the FlagTransactionPosted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagTransactionPosted

`func (o *ConsensusWorkflowMetrics) SetFlagTransactionPosted(v bool)`

SetFlagTransactionPosted sets FlagTransactionPosted field to given value.


### GetFlagTransactionSeen

`func (o *ConsensusWorkflowMetrics) GetFlagTransactionSeen() bool`

GetFlagTransactionSeen returns the FlagTransactionSeen field if non-nil, zero value otherwise.

### GetFlagTransactionSeenOk

`func (o *ConsensusWorkflowMetrics) GetFlagTransactionSeenOk() (*bool, bool)`

GetFlagTransactionSeenOk returns a tuple with the FlagTransactionSeen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagTransactionSeen

`func (o *ConsensusWorkflowMetrics) SetFlagTransactionSeen(v bool)`

SetFlagTransactionSeen sets FlagTransactionSeen field to given value.


### GetFlagVMResultSigned

`func (o *ConsensusWorkflowMetrics) GetFlagVMResultSigned() bool`

GetFlagVMResultSigned returns the FlagVMResultSigned field if non-nil, zero value otherwise.

### GetFlagVMResultSignedOk

`func (o *ConsensusWorkflowMetrics) GetFlagVMResultSignedOk() (*bool, bool)`

GetFlagVMResultSignedOk returns a tuple with the FlagVMResultSigned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagVMResultSigned

`func (o *ConsensusWorkflowMetrics) SetFlagVMResultSigned(v bool)`

SetFlagVMResultSigned sets FlagVMResultSigned field to given value.


### GetFlagVMStarted

`func (o *ConsensusWorkflowMetrics) GetFlagVMStarted() bool`

GetFlagVMStarted returns the FlagVMStarted field if non-nil, zero value otherwise.

### GetFlagVMStartedOk

`func (o *ConsensusWorkflowMetrics) GetFlagVMStartedOk() (*bool, bool)`

GetFlagVMStartedOk returns a tuple with the FlagVMStarted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagVMStarted

`func (o *ConsensusWorkflowMetrics) SetFlagVMStarted(v bool)`

SetFlagVMStarted sets FlagVMStarted field to given value.


### GetTimeBatchProposalSent

`func (o *ConsensusWorkflowMetrics) GetTimeBatchProposalSent() time.Time`

GetTimeBatchProposalSent returns the TimeBatchProposalSent field if non-nil, zero value otherwise.

### GetTimeBatchProposalSentOk

`func (o *ConsensusWorkflowMetrics) GetTimeBatchProposalSentOk() (*time.Time, bool)`

GetTimeBatchProposalSentOk returns a tuple with the TimeBatchProposalSent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeBatchProposalSent

`func (o *ConsensusWorkflowMetrics) SetTimeBatchProposalSent(v time.Time)`

SetTimeBatchProposalSent sets TimeBatchProposalSent field to given value.


### GetTimeCompleted

`func (o *ConsensusWorkflowMetrics) GetTimeCompleted() time.Time`

GetTimeCompleted returns the TimeCompleted field if non-nil, zero value otherwise.

### GetTimeCompletedOk

`func (o *ConsensusWorkflowMetrics) GetTimeCompletedOk() (*time.Time, bool)`

GetTimeCompletedOk returns a tuple with the TimeCompleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeCompleted

`func (o *ConsensusWorkflowMetrics) SetTimeCompleted(v time.Time)`

SetTimeCompleted sets TimeCompleted field to given value.


### GetTimeConsensusBatchKnown

`func (o *ConsensusWorkflowMetrics) GetTimeConsensusBatchKnown() time.Time`

GetTimeConsensusBatchKnown returns the TimeConsensusBatchKnown field if non-nil, zero value otherwise.

### GetTimeConsensusBatchKnownOk

`func (o *ConsensusWorkflowMetrics) GetTimeConsensusBatchKnownOk() (*time.Time, bool)`

GetTimeConsensusBatchKnownOk returns a tuple with the TimeConsensusBatchKnown field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeConsensusBatchKnown

`func (o *ConsensusWorkflowMetrics) SetTimeConsensusBatchKnown(v time.Time)`

SetTimeConsensusBatchKnown sets TimeConsensusBatchKnown field to given value.


### GetTimeTransactionFinalized

`func (o *ConsensusWorkflowMetrics) GetTimeTransactionFinalized() time.Time`

GetTimeTransactionFinalized returns the TimeTransactionFinalized field if non-nil, zero value otherwise.

### GetTimeTransactionFinalizedOk

`func (o *ConsensusWorkflowMetrics) GetTimeTransactionFinalizedOk() (*time.Time, bool)`

GetTimeTransactionFinalizedOk returns a tuple with the TimeTransactionFinalized field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeTransactionFinalized

`func (o *ConsensusWorkflowMetrics) SetTimeTransactionFinalized(v time.Time)`

SetTimeTransactionFinalized sets TimeTransactionFinalized field to given value.


### GetTimeTransactionPosted

`func (o *ConsensusWorkflowMetrics) GetTimeTransactionPosted() time.Time`

GetTimeTransactionPosted returns the TimeTransactionPosted field if non-nil, zero value otherwise.

### GetTimeTransactionPostedOk

`func (o *ConsensusWorkflowMetrics) GetTimeTransactionPostedOk() (*time.Time, bool)`

GetTimeTransactionPostedOk returns a tuple with the TimeTransactionPosted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeTransactionPosted

`func (o *ConsensusWorkflowMetrics) SetTimeTransactionPosted(v time.Time)`

SetTimeTransactionPosted sets TimeTransactionPosted field to given value.


### GetTimeTransactionSeen

`func (o *ConsensusWorkflowMetrics) GetTimeTransactionSeen() time.Time`

GetTimeTransactionSeen returns the TimeTransactionSeen field if non-nil, zero value otherwise.

### GetTimeTransactionSeenOk

`func (o *ConsensusWorkflowMetrics) GetTimeTransactionSeenOk() (*time.Time, bool)`

GetTimeTransactionSeenOk returns a tuple with the TimeTransactionSeen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeTransactionSeen

`func (o *ConsensusWorkflowMetrics) SetTimeTransactionSeen(v time.Time)`

SetTimeTransactionSeen sets TimeTransactionSeen field to given value.


### GetTimeVMResultSigned

`func (o *ConsensusWorkflowMetrics) GetTimeVMResultSigned() time.Time`

GetTimeVMResultSigned returns the TimeVMResultSigned field if non-nil, zero value otherwise.

### GetTimeVMResultSignedOk

`func (o *ConsensusWorkflowMetrics) GetTimeVMResultSignedOk() (*time.Time, bool)`

GetTimeVMResultSignedOk returns a tuple with the TimeVMResultSigned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeVMResultSigned

`func (o *ConsensusWorkflowMetrics) SetTimeVMResultSigned(v time.Time)`

SetTimeVMResultSigned sets TimeVMResultSigned field to given value.


### GetTimeVMStarted

`func (o *ConsensusWorkflowMetrics) GetTimeVMStarted() time.Time`

GetTimeVMStarted returns the TimeVMStarted field if non-nil, zero value otherwise.

### GetTimeVMStartedOk

`func (o *ConsensusWorkflowMetrics) GetTimeVMStartedOk() (*time.Time, bool)`

GetTimeVMStartedOk returns a tuple with the TimeVMStarted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeVMStarted

`func (o *ConsensusWorkflowMetrics) SetTimeVMStarted(v time.Time)`

SetTimeVMStarted sets TimeVMStarted field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


