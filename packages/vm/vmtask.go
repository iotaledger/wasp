package vm

import (
	"time"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type VMRunner interface {
	Run(task *VMTask)
}

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
// It is assumed that all requests/inputs are unlock-able by alaisAddress of provided ChainInput
// at timestamp = Timestamp + len(Requests) nanoseconds
type VMTask struct {
	ACSSessionID             uint64
	Processors               *processors.ProcessorCache
	ChainInput               *ledgerstate.AliasOutput
	VirtualState             state.VirtualState // in/out  Return uncommitted updated virtual state
	SolidStateBaseline       *coreutil.SolidStateBaseline
	Requests                 []coretypes.Request
	Timestamp                time.Time
	Entropy                  hashing.HashValue
	ValidatorFeeTarget       coretypes.AgentID
	Log                      *logger.Logger
	OnFinish                 func(callResult dict.Dict, callError error, vmError error)
	ResultTransactionEssence *ledgerstate.TransactionEssence
}
