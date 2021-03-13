package vm

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction_old"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type RequestRefWithFreeTokens struct {
	sctransaction_old.RequestRef
	FreeTokens coretypes.ColoredBalancesOld
}

// task context (for batch of requests)
type VMTask struct {
	Processors *processors.ProcessorCache
	// inputs (immutable)
	ChainID coretypes.ChainID
	Color   balance.Color
	// deterministic source of entropy
	Entropy            hashing.HashValue
	Balances           map[valuetransaction.ID][]*balance.Balance
	ValidatorFeeTarget coretypes.AgentID
	Requests           []RequestRefWithFreeTokens
	Timestamp          int64
	VirtualState       state.VirtualState // input immutable
	Log                *logger.Logger
	// call when finished
	OnFinish func(callResult dict.Dict, callError error, vmError error)
	// outputs
	ResultTransaction *sctransaction_old.TransactionEssence
	ResultBlock       state.Block
}

// BatchHash is used to uniquely identify the VM task
func BatchHash(reqids []coretypes.RequestID, ts int64, leaderIndex uint16) hashing.HashValue {
	var buf bytes.Buffer
	for i := range reqids {
		buf.Write(reqids[i][:])
	}
	_ = util.WriteInt64(&buf, ts)
	_ = util.WriteUint16(&buf, leaderIndex)

	return hashing.HashData(buf.Bytes())
}
