// Package vm defines the types required for the vm
package vm

import (
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
// It is assumed that all requests/inputs are unlock-able by aliasAddress of provided AnchorOutput
// at timestamp = Timestamp + len(Requests) nanoseconds
type VMTask struct {
	Processors *processors.Config
	Anchor     *isc.StateAnchor
	// GasCoin is allowed to be nil iif EstimateGasMode == true || EVMTracer != nil,
	// in which case no PTB will be produced.
	GasCoin            *coin.CoinWithRef
	Store              state.Store
	Requests           []isc.Request
	Timestamp          time.Time
	Entropy            hashing.HashValue
	ValidatorFeeTarget isc.AgentID
	L1Params           *parameters.L1Params
	// If EstimateGasMode is enabled, signature and nonce checks will be skipped
	EstimateGasMode      bool
	EVMTracer            *tracers.Tracer
	EnableGasBurnLogging bool // for testing and Solo only

	Migrations *migrations.MigrationScheme // for testing and Solo only

	Log log.Logger
}

type VMTaskResult struct {
	Task *VMTask

	// StateDraft is the uncommitted state resulting from the execution of the requests
	StateDraft state.StateDraft
	// PTB is the ProgrammableTransaction to be sent to L1 for the next anchor
	// transition, or nil if the task does not produce a normal block
	UnsignedTransaction iotago.ProgrammableTransaction
	StateMetadata       []byte
	// RequestResults contains one result for each non-skipped request
	RequestResults []*RequestResult
}

type RequestResult struct {
	// Request is the corresponding request in the task
	Request isc.Request
	// Return is the return value of the call
	Return isc.CallArguments
	// Receipt is the receipt produced after executing the request
	Receipt *blocklog.RequestReceipt
}

func (task *VMTask) WillProduceBlock() bool {
	return !task.EstimateGasMode && task.EVMTracer == nil
}

func (task *VMTask) FinalStateTimestamp() time.Time {
	return task.Timestamp.Add(time.Duration(len(task.Requests)+1) * time.Nanosecond)
}
