package vm

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"time"
)

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
// It is assumed that all requests/inputs are unlock-able by alaisAddress of provided ChainInput
// at timestamp = Timestamp + len(Requests) nanoseconds
type VMTask struct {
	Processors         *processors.ProcessorCache
	ChainInput         *ledgerstate.AliasOutput
	VirtualState       state.VirtualState
	Requests           []coretypes.Request
	Timestamp          time.Time
	Entropy            hashing.HashValue
	ValidatorFeeTarget coretypes.AgentID
	Log                *logger.Logger
	// call when finished
	OnFinish func(callResult dict.Dict, callError error, vmError error)
	// result
	ResultTransaction *ledgerstate.TransactionEssence
	ResultBlock       state.Block
}

// BatchHash is used to uniquely identify the VM task
func BatchHash(reqids []coretypes.RequestID, ts time.Time, leaderIndex uint16) hashing.HashValue {
	var buf bytes.Buffer
	for i := range reqids {
		buf.Write(reqids[i][:])
	}
	_ = util.WriteInt64(&buf, ts.UnixNano())
	_ = util.WriteUint16(&buf, leaderIndex)

	return hashing.HashData(buf.Bytes())
}
