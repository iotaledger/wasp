package vm

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
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
	Processors               *processors.Cache
	ChainInput               *ledgerstate.AliasOutput
	VirtualState             state.VirtualState // in/out  Return uncommitted updated virtual state
	SolidStateBaseline       coreutil.StateBaseline
	Requests                 []iscp.Request
	Timestamp                time.Time
	Entropy                  hashing.HashValue
	ValidatorFeeTarget       iscp.AgentID
	Log                      *logger.Logger
	OnFinish                 func(callResult dict.Dict, callError error, vmError error)
	ResultTransactionEssence *ledgerstate.TransactionEssence // if not nil it is a normal block
	RotationAddress          ledgerstate.Address             // if not nil, it is a rotation
}
