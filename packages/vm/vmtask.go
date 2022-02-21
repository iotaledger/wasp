package vm

import (
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type VMRunner interface {
	Run(task *VMTask)
}

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
// It is assumed that all requests/inputs are unlock-able by aliasAddress of provided AnchorOutput
// at timestamp = Timestamp + len(Requests) nanoseconds
type VMTask struct {
	// INPUTS:

	ACSSessionID       uint64
	Processors         *processors.Cache
	AnchorOutput       *iotago.AliasOutput
	AnchorOutputID     iotago.OutputID
	L1Params           *parameters.L1
	SolidStateBaseline coreutil.StateBaseline
	Requests           []iscp.Request
	TimeAssumption     iscp.TimeData
	Entropy            hashing.HashValue
	ValidatorFeeTarget *iscp.AgentID
	// If EstimateGasMode is enabled, gas fee will be calculated but not charged
	EstimateGasMode      bool
	EnableGasBurnLogging bool // for testing and Solo only

	// INPUTS_OUTPUTS:

	// VirtualStateAccess is the initial state of the chain, which is also
	// mutated during the execution of the task
	VirtualStateAccess state.VirtualStateAccess
	Log                *logger.Logger

	// OUTPUTS:

	// RotationAddress is the next address after a rotation, or nil if there is no rotation
	RotationAddress iotago.Address
	// TransactionEssence is the transaction essence for the next block,
	// or nil if the task does not produce a normal block
	ResultTransactionEssence *iotago.TransactionEssence
	// ResultInputsCommitment is the inputs commitment necessary to sign the ResultTransactionEssence
	ResultInputsCommitment []byte
	// Results contains one result for each non-skipped request
	Results []*RequestResult
	// If not nil, VMError is a fatal error that prevented the execution of the task
	VMError error
}

type RequestResult struct {
	// Request is the corresponding request in the task
	Request iscp.Request
	// Return is the return value of the call
	Return dict.Dict
	// Error is the error produced by the call, if any
	Error error
	// Receipt is the receipt produced after executing the request
	Receipt *blocklog.RequestReceipt
}

func (task *VMTask) GetProcessedRequestIDs() []iscp.RequestID {
	ret := make([]iscp.RequestID, len(task.Results))
	for i, res := range task.Results {
		ret[i] = res.Request.ID()
	}
	return ret
}
