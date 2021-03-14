package vm

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"time"
)

// VMTask is task context (for batch of requests). It is used to pass parameters and take results
type VMTask struct {
	Processors         *processors.ProcessorCache
	ChainInput         *ledgerstate.ChainOutput
	Requests           []*sctransaction.Request
	Timestamp          time.Time
	VirtualState       state.VirtualState // input immutable
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
func BatchHash(reqids []ledgerstate.OutputID, ts time.Time, leaderIndex uint16) hashing.HashValue {
	var buf bytes.Buffer
	for i := range reqids {
		buf.Write(reqids[i][:])
	}
	_ = util.WriteInt64(&buf, ts.UnixNano())
	_ = util.WriteUint16(&buf, leaderIndex)

	return hashing.HashData(buf.Bytes())
}
