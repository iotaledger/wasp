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
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
// It is assumed that all requests/inputs are unlock-able by aliasAddress of provided AnchorOutput
// at timestamp = Timestamp + len(Requests) nanoseconds
type VMTask struct {
	Processors         *processors.Cache
	AnchorOutput       *iotago.AliasOutput
	AnchorOutputID     iotago.OutputID
	Store              state.Store
	Requests           []isc.Request
	TimeAssumption     time.Time
	Entropy            hashing.HashValue
	ValidatorFeeTarget isc.AgentID
	// If EstimateGasMode is enabled, gas fee will be calculated but not charged
	EstimateGasMode bool
	// If EVMTracer is set, all requests will be executed normally up until the EVM
	// tx with the given index, which will then be executed with the given tracer.
	EVMTracer            *isc.EVMTracer
	EnableGasBurnLogging bool // for testing and Solo only

	MigrationsOverride *migrations.MigrationScheme // for testing and Solo only

	Log *logger.Logger
}

type VMTaskResult struct {
	Task *VMTask

	// StateDraft is the uncommitted state resulting from the execution of the requests
	StateDraft state.StateDraft
	// RotationAddress is the next address after a rotation, or nil if there is no rotation
	RotationAddress iotago.Address
	// TransactionEssence is the transaction essence for the next block,
	// or nil if the task does not produce a normal block
	TransactionEssence *iotago.TransactionEssence
	// InputsCommitment is the inputs commitment necessary to sign the ResultTransactionEssence
	InputsCommitment []byte
	// RequestResults contains one result for each non-skipped request
	RequestResults []*RequestResult
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

func (task *VMTask) FinalStateTimestamp() time.Time {
	return task.TimeAssumption.Add(time.Duration(len(task.Requests)+1) * time.Nanosecond)
}
