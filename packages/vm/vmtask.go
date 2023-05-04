package vm

import (
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type VMRunner interface {
	Run(task *VMTask) error
}

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
// It is assumed that all requests/inputs are unlock-able by aliasAddress of provided AnchorOutput
// at timestamp = Timestamp + len(Requests) nanoseconds
type VMTask struct {
	// INPUTS:

	Processors                 *processors.Cache
	AnchorOutput               *iotago.AliasOutput
	AnchorOutputID             iotago.OutputID
	AnchorOutputStorageDeposit uint64 // will be filled by vmcontext
	Store                      state.Store
	Requests                   []isc.Request
	UnprocessableToRetry       []isc.Request
	TimeAssumption             time.Time
	Entropy                    hashing.HashValue
	ValidatorFeeTarget         isc.AgentID
	// If EstimateGasMode is enabled, gas fee will be calculated but not charged
	EstimateGasMode bool
	// If EVMtrace is set, all requests will be executed normally up until the EVM
	// tx with the given index, which will then be executed with the given tracer.
	EVMTracer            *isc.EVMTracer
	EnableGasBurnLogging bool // for testing and Solo only

	// INPUTS_OUTPUTS:

	Log *logger.Logger

	// OUTPUTS:

	// the uncommitted state resulting from the execution of the requests
	StateDraft state.StateDraft
	// RotationAddress is the next address after a rotation, or nil if there is no rotation
	RotationAddress iotago.Address
	// TransactionEssence is the transaction essence for the next block,
	// or nil if the task does not produce a normal block
	ResultTransactionEssence *iotago.TransactionEssence
	// ResultInputsCommitment is the inputs commitment necessary to sign the ResultTransactionEssence
	ResultInputsCommitment []byte
	// Results contains one result for each non-skipped request
	Results []*RequestResult
	// If maintenance mode is enabled, only requests to the governance contract will be executed
	MaintenanceModeEnabled bool
}

type RequestResult struct {
	// Request is the corresponding request in the task
	Request isc.Request
	// Return is the return value of the call
	Return dict.Dict
	// Receipt is the receipt produced after executing the request
	Receipt *blocklog.RequestReceipt
}

func (task *VMTask) WillProduceBlock() bool {
	return !task.EstimateGasMode && task.EVMTracer == nil
}
