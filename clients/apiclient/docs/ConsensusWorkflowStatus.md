# ConsensusWorkflowStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CurrentStateIndex** | Pointer to **int32** | Shows current state index of the consensus | [optional] 
**FlagBatchProposalSent** | Pointer to **bool** | Shows if batch proposal is sent out in current consensus iteration | [optional] 
**FlagConsensusBatchKnown** | Pointer to **bool** | Shows if consensus on batch is reached and known in current consensus iteration | [optional] 
**FlagInProgress** | Pointer to **bool** | Shows if consensus algorithm is still not completed in current consensus iteration | [optional] 
**FlagStateReceived** | Pointer to **bool** | Shows if state output is received in current consensus iteration | [optional] 
**FlagTransactionFinalized** | Pointer to **bool** | Shows if consensus on transaction is reached in current consensus iteration | [optional] 
**FlagTransactionPosted** | Pointer to **bool** | Shows if transaction is posted to L1 in current consensus iteration | [optional] 
**FlagTransactionSeen** | Pointer to **bool** | Shows if L1 reported that it has seen the transaction of current consensus iteration | [optional] 
**FlagVMResultSigned** | Pointer to **bool** | Shows if virtual machine has returned its results in current consensus iteration | [optional] 
**FlagVMStarted** | Pointer to **bool** | Shows if virtual machine is started in current consensus iteration | [optional] 
**TimeBatchProposalSent** | Pointer to **time.Time** | Shows when batch proposal was last sent out in current consensus iteration | [optional] 
**TimeCompleted** | Pointer to **time.Time** | Shows when algorithm was last completed in current consensus iteration | [optional] 
**TimeConsensusBatchKnown** | Pointer to **time.Time** | Shows when ACS results of consensus on batch was last received in current consensus iteration | [optional] 
**TimeTransactionFinalized** | Pointer to **time.Time** | Shows when algorithm last noted that all the data for consensus on transaction had been received in current consensus iteration | [optional] 
**TimeTransactionPosted** | Pointer to **time.Time** | Shows when transaction was last posted to L1 in current consensus iteration | [optional] 
**TimeTransactionSeen** | Pointer to **time.Time** | Shows when algorithm last noted that transaction hadd been seen by L1 in current consensus iteration | [optional] 
**TimeVMResultSigned** | Pointer to **time.Time** | Shows when virtual machine results were last received and signed in current consensus iteration | [optional] 
**TimeVMStarted** | Pointer to **time.Time** | Shows when virtual machine was last started in current consensus iteration | [optional] 

## Methods

### NewConsensusWorkflowStatus

`func NewConsensusWorkflowStatus() *ConsensusWorkflowStatus`

NewConsensusWorkflowStatus instantiates a new ConsensusWorkflowStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConsensusWorkflowStatusWithDefaults

`func NewConsensusWorkflowStatusWithDefaults() *ConsensusWorkflowStatus`

NewConsensusWorkflowStatusWithDefaults instantiates a new ConsensusWorkflowStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCurrentStateIndex

`func (o *ConsensusWorkflowStatus) GetCurrentStateIndex() int32`

GetCurrentStateIndex returns the CurrentStateIndex field if non-nil, zero value otherwise.

### GetCurrentStateIndexOk

`func (o *ConsensusWorkflowStatus) GetCurrentStateIndexOk() (*int32, bool)`

GetCurrentStateIndexOk returns a tuple with the CurrentStateIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentStateIndex

`func (o *ConsensusWorkflowStatus) SetCurrentStateIndex(v int32)`

SetCurrentStateIndex sets CurrentStateIndex field to given value.

### HasCurrentStateIndex

`func (o *ConsensusWorkflowStatus) HasCurrentStateIndex() bool`

HasCurrentStateIndex returns a boolean if a field has been set.

### GetFlagBatchProposalSent

`func (o *ConsensusWorkflowStatus) GetFlagBatchProposalSent() bool`

GetFlagBatchProposalSent returns the FlagBatchProposalSent field if non-nil, zero value otherwise.

### GetFlagBatchProposalSentOk

`func (o *ConsensusWorkflowStatus) GetFlagBatchProposalSentOk() (*bool, bool)`

GetFlagBatchProposalSentOk returns a tuple with the FlagBatchProposalSent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagBatchProposalSent

`func (o *ConsensusWorkflowStatus) SetFlagBatchProposalSent(v bool)`

SetFlagBatchProposalSent sets FlagBatchProposalSent field to given value.

### HasFlagBatchProposalSent

`func (o *ConsensusWorkflowStatus) HasFlagBatchProposalSent() bool`

HasFlagBatchProposalSent returns a boolean if a field has been set.

### GetFlagConsensusBatchKnown

`func (o *ConsensusWorkflowStatus) GetFlagConsensusBatchKnown() bool`

GetFlagConsensusBatchKnown returns the FlagConsensusBatchKnown field if non-nil, zero value otherwise.

### GetFlagConsensusBatchKnownOk

`func (o *ConsensusWorkflowStatus) GetFlagConsensusBatchKnownOk() (*bool, bool)`

GetFlagConsensusBatchKnownOk returns a tuple with the FlagConsensusBatchKnown field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagConsensusBatchKnown

`func (o *ConsensusWorkflowStatus) SetFlagConsensusBatchKnown(v bool)`

SetFlagConsensusBatchKnown sets FlagConsensusBatchKnown field to given value.

### HasFlagConsensusBatchKnown

`func (o *ConsensusWorkflowStatus) HasFlagConsensusBatchKnown() bool`

HasFlagConsensusBatchKnown returns a boolean if a field has been set.

### GetFlagInProgress

`func (o *ConsensusWorkflowStatus) GetFlagInProgress() bool`

GetFlagInProgress returns the FlagInProgress field if non-nil, zero value otherwise.

### GetFlagInProgressOk

`func (o *ConsensusWorkflowStatus) GetFlagInProgressOk() (*bool, bool)`

GetFlagInProgressOk returns a tuple with the FlagInProgress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagInProgress

`func (o *ConsensusWorkflowStatus) SetFlagInProgress(v bool)`

SetFlagInProgress sets FlagInProgress field to given value.

### HasFlagInProgress

`func (o *ConsensusWorkflowStatus) HasFlagInProgress() bool`

HasFlagInProgress returns a boolean if a field has been set.

### GetFlagStateReceived

`func (o *ConsensusWorkflowStatus) GetFlagStateReceived() bool`

GetFlagStateReceived returns the FlagStateReceived field if non-nil, zero value otherwise.

### GetFlagStateReceivedOk

`func (o *ConsensusWorkflowStatus) GetFlagStateReceivedOk() (*bool, bool)`

GetFlagStateReceivedOk returns a tuple with the FlagStateReceived field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagStateReceived

`func (o *ConsensusWorkflowStatus) SetFlagStateReceived(v bool)`

SetFlagStateReceived sets FlagStateReceived field to given value.

### HasFlagStateReceived

`func (o *ConsensusWorkflowStatus) HasFlagStateReceived() bool`

HasFlagStateReceived returns a boolean if a field has been set.

### GetFlagTransactionFinalized

`func (o *ConsensusWorkflowStatus) GetFlagTransactionFinalized() bool`

GetFlagTransactionFinalized returns the FlagTransactionFinalized field if non-nil, zero value otherwise.

### GetFlagTransactionFinalizedOk

`func (o *ConsensusWorkflowStatus) GetFlagTransactionFinalizedOk() (*bool, bool)`

GetFlagTransactionFinalizedOk returns a tuple with the FlagTransactionFinalized field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagTransactionFinalized

`func (o *ConsensusWorkflowStatus) SetFlagTransactionFinalized(v bool)`

SetFlagTransactionFinalized sets FlagTransactionFinalized field to given value.

### HasFlagTransactionFinalized

`func (o *ConsensusWorkflowStatus) HasFlagTransactionFinalized() bool`

HasFlagTransactionFinalized returns a boolean if a field has been set.

### GetFlagTransactionPosted

`func (o *ConsensusWorkflowStatus) GetFlagTransactionPosted() bool`

GetFlagTransactionPosted returns the FlagTransactionPosted field if non-nil, zero value otherwise.

### GetFlagTransactionPostedOk

`func (o *ConsensusWorkflowStatus) GetFlagTransactionPostedOk() (*bool, bool)`

GetFlagTransactionPostedOk returns a tuple with the FlagTransactionPosted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagTransactionPosted

`func (o *ConsensusWorkflowStatus) SetFlagTransactionPosted(v bool)`

SetFlagTransactionPosted sets FlagTransactionPosted field to given value.

### HasFlagTransactionPosted

`func (o *ConsensusWorkflowStatus) HasFlagTransactionPosted() bool`

HasFlagTransactionPosted returns a boolean if a field has been set.

### GetFlagTransactionSeen

`func (o *ConsensusWorkflowStatus) GetFlagTransactionSeen() bool`

GetFlagTransactionSeen returns the FlagTransactionSeen field if non-nil, zero value otherwise.

### GetFlagTransactionSeenOk

`func (o *ConsensusWorkflowStatus) GetFlagTransactionSeenOk() (*bool, bool)`

GetFlagTransactionSeenOk returns a tuple with the FlagTransactionSeen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagTransactionSeen

`func (o *ConsensusWorkflowStatus) SetFlagTransactionSeen(v bool)`

SetFlagTransactionSeen sets FlagTransactionSeen field to given value.

### HasFlagTransactionSeen

`func (o *ConsensusWorkflowStatus) HasFlagTransactionSeen() bool`

HasFlagTransactionSeen returns a boolean if a field has been set.

### GetFlagVMResultSigned

`func (o *ConsensusWorkflowStatus) GetFlagVMResultSigned() bool`

GetFlagVMResultSigned returns the FlagVMResultSigned field if non-nil, zero value otherwise.

### GetFlagVMResultSignedOk

`func (o *ConsensusWorkflowStatus) GetFlagVMResultSignedOk() (*bool, bool)`

GetFlagVMResultSignedOk returns a tuple with the FlagVMResultSigned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagVMResultSigned

`func (o *ConsensusWorkflowStatus) SetFlagVMResultSigned(v bool)`

SetFlagVMResultSigned sets FlagVMResultSigned field to given value.

### HasFlagVMResultSigned

`func (o *ConsensusWorkflowStatus) HasFlagVMResultSigned() bool`

HasFlagVMResultSigned returns a boolean if a field has been set.

### GetFlagVMStarted

`func (o *ConsensusWorkflowStatus) GetFlagVMStarted() bool`

GetFlagVMStarted returns the FlagVMStarted field if non-nil, zero value otherwise.

### GetFlagVMStartedOk

`func (o *ConsensusWorkflowStatus) GetFlagVMStartedOk() (*bool, bool)`

GetFlagVMStartedOk returns a tuple with the FlagVMStarted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlagVMStarted

`func (o *ConsensusWorkflowStatus) SetFlagVMStarted(v bool)`

SetFlagVMStarted sets FlagVMStarted field to given value.

### HasFlagVMStarted

`func (o *ConsensusWorkflowStatus) HasFlagVMStarted() bool`

HasFlagVMStarted returns a boolean if a field has been set.

### GetTimeBatchProposalSent

`func (o *ConsensusWorkflowStatus) GetTimeBatchProposalSent() time.Time`

GetTimeBatchProposalSent returns the TimeBatchProposalSent field if non-nil, zero value otherwise.

### GetTimeBatchProposalSentOk

`func (o *ConsensusWorkflowStatus) GetTimeBatchProposalSentOk() (*time.Time, bool)`

GetTimeBatchProposalSentOk returns a tuple with the TimeBatchProposalSent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeBatchProposalSent

`func (o *ConsensusWorkflowStatus) SetTimeBatchProposalSent(v time.Time)`

SetTimeBatchProposalSent sets TimeBatchProposalSent field to given value.

### HasTimeBatchProposalSent

`func (o *ConsensusWorkflowStatus) HasTimeBatchProposalSent() bool`

HasTimeBatchProposalSent returns a boolean if a field has been set.

### GetTimeCompleted

`func (o *ConsensusWorkflowStatus) GetTimeCompleted() time.Time`

GetTimeCompleted returns the TimeCompleted field if non-nil, zero value otherwise.

### GetTimeCompletedOk

`func (o *ConsensusWorkflowStatus) GetTimeCompletedOk() (*time.Time, bool)`

GetTimeCompletedOk returns a tuple with the TimeCompleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeCompleted

`func (o *ConsensusWorkflowStatus) SetTimeCompleted(v time.Time)`

SetTimeCompleted sets TimeCompleted field to given value.

### HasTimeCompleted

`func (o *ConsensusWorkflowStatus) HasTimeCompleted() bool`

HasTimeCompleted returns a boolean if a field has been set.

### GetTimeConsensusBatchKnown

`func (o *ConsensusWorkflowStatus) GetTimeConsensusBatchKnown() time.Time`

GetTimeConsensusBatchKnown returns the TimeConsensusBatchKnown field if non-nil, zero value otherwise.

### GetTimeConsensusBatchKnownOk

`func (o *ConsensusWorkflowStatus) GetTimeConsensusBatchKnownOk() (*time.Time, bool)`

GetTimeConsensusBatchKnownOk returns a tuple with the TimeConsensusBatchKnown field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeConsensusBatchKnown

`func (o *ConsensusWorkflowStatus) SetTimeConsensusBatchKnown(v time.Time)`

SetTimeConsensusBatchKnown sets TimeConsensusBatchKnown field to given value.

### HasTimeConsensusBatchKnown

`func (o *ConsensusWorkflowStatus) HasTimeConsensusBatchKnown() bool`

HasTimeConsensusBatchKnown returns a boolean if a field has been set.

### GetTimeTransactionFinalized

`func (o *ConsensusWorkflowStatus) GetTimeTransactionFinalized() time.Time`

GetTimeTransactionFinalized returns the TimeTransactionFinalized field if non-nil, zero value otherwise.

### GetTimeTransactionFinalizedOk

`func (o *ConsensusWorkflowStatus) GetTimeTransactionFinalizedOk() (*time.Time, bool)`

GetTimeTransactionFinalizedOk returns a tuple with the TimeTransactionFinalized field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeTransactionFinalized

`func (o *ConsensusWorkflowStatus) SetTimeTransactionFinalized(v time.Time)`

SetTimeTransactionFinalized sets TimeTransactionFinalized field to given value.

### HasTimeTransactionFinalized

`func (o *ConsensusWorkflowStatus) HasTimeTransactionFinalized() bool`

HasTimeTransactionFinalized returns a boolean if a field has been set.

### GetTimeTransactionPosted

`func (o *ConsensusWorkflowStatus) GetTimeTransactionPosted() time.Time`

GetTimeTransactionPosted returns the TimeTransactionPosted field if non-nil, zero value otherwise.

### GetTimeTransactionPostedOk

`func (o *ConsensusWorkflowStatus) GetTimeTransactionPostedOk() (*time.Time, bool)`

GetTimeTransactionPostedOk returns a tuple with the TimeTransactionPosted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeTransactionPosted

`func (o *ConsensusWorkflowStatus) SetTimeTransactionPosted(v time.Time)`

SetTimeTransactionPosted sets TimeTransactionPosted field to given value.

### HasTimeTransactionPosted

`func (o *ConsensusWorkflowStatus) HasTimeTransactionPosted() bool`

HasTimeTransactionPosted returns a boolean if a field has been set.

### GetTimeTransactionSeen

`func (o *ConsensusWorkflowStatus) GetTimeTransactionSeen() time.Time`

GetTimeTransactionSeen returns the TimeTransactionSeen field if non-nil, zero value otherwise.

### GetTimeTransactionSeenOk

`func (o *ConsensusWorkflowStatus) GetTimeTransactionSeenOk() (*time.Time, bool)`

GetTimeTransactionSeenOk returns a tuple with the TimeTransactionSeen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeTransactionSeen

`func (o *ConsensusWorkflowStatus) SetTimeTransactionSeen(v time.Time)`

SetTimeTransactionSeen sets TimeTransactionSeen field to given value.

### HasTimeTransactionSeen

`func (o *ConsensusWorkflowStatus) HasTimeTransactionSeen() bool`

HasTimeTransactionSeen returns a boolean if a field has been set.

### GetTimeVMResultSigned

`func (o *ConsensusWorkflowStatus) GetTimeVMResultSigned() time.Time`

GetTimeVMResultSigned returns the TimeVMResultSigned field if non-nil, zero value otherwise.

### GetTimeVMResultSignedOk

`func (o *ConsensusWorkflowStatus) GetTimeVMResultSignedOk() (*time.Time, bool)`

GetTimeVMResultSignedOk returns a tuple with the TimeVMResultSigned field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeVMResultSigned

`func (o *ConsensusWorkflowStatus) SetTimeVMResultSigned(v time.Time)`

SetTimeVMResultSigned sets TimeVMResultSigned field to given value.

### HasTimeVMResultSigned

`func (o *ConsensusWorkflowStatus) HasTimeVMResultSigned() bool`

HasTimeVMResultSigned returns a boolean if a field has been set.

### GetTimeVMStarted

`func (o *ConsensusWorkflowStatus) GetTimeVMStarted() time.Time`

GetTimeVMStarted returns the TimeVMStarted field if non-nil, zero value otherwise.

### GetTimeVMStartedOk

`func (o *ConsensusWorkflowStatus) GetTimeVMStartedOk() (*time.Time, bool)`

GetTimeVMStartedOk returns a tuple with the TimeVMStarted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeVMStarted

`func (o *ConsensusWorkflowStatus) SetTimeVMStarted(v time.Time)`

SetTimeVMStarted sets TimeVMStarted field to given value.

### HasTimeVMStarted

`func (o *ConsensusWorkflowStatus) HasTimeVMStarted() bool`

HasTimeVMStarted returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


