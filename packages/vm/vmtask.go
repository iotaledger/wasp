package vm

import (
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type VMRunner interface {
	Run(task *VMTask)
}

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
// It is assumed that all requests/inputs are unlock-able by aliasAddress of provided AnchorOutput
// at timestamp = Timestamp + len(Requests) nanoseconds
type VMTask struct {
	ACSSessionID             uint64
	Processors               *processors.Cache
	AnchorOutput             *iotago.AliasOutput
	AnchorOutputID           iotago.UTXOInput
	RentStructure            *iotago.RentStructure
	VirtualStateAccess       state.VirtualStateAccess
	SolidStateBaseline       coreutil.StateBaseline
	Requests                 []iscp.RequestData
	ProcessedRequestsCount   uint16
	TimeAssumption           iscp.TimeData
	Entropy                  hashing.HashValue
	ValidatorFeeTarget       *iscp.AgentID
	Log                      *logger.Logger
	OnFinish                 func(callResult dict.Dict, callError error, vmError error)
	ResultTransactionEssence *iotago.TransactionEssence // if not nil it is a normal block
	RotationAddress          iotago.Address             // if not nil, it is a rotation
	StartTime                time.Time
}
