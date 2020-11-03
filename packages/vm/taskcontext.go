package vm

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// task context (for batch of requests)
type VMTask struct {
	Processors *processors.ProcessorCache
	// inputs (immutable)
	LeaderPeerIndex uint16
	ProgramHash     hashing.HashValue
	ChainID         coretypes.ChainID
	Color           balance.Color
	// deterministic source of entropy (pseudorandom, unpredictable for parties)
	Entropy       hashing.HashValue
	Balances      map[valuetransaction.ID][]*balance.Balance
	OwnerAddress  address.Address
	RewardAddress address.Address
	MinimumReward int64
	Requests      []sctransaction.RequestRef
	Timestamp     int64
	VirtualState  state.VirtualState // input immutable
	Log           *logger.Logger
	// call when finished
	OnFinish func(error)
	// outputs
	ResultTransaction *sctransaction.Transaction
	ResultBlock       state.Block
}

// BatchHash is used to uniquely identify the VM task
func BatchHash(reqids []coretypes.RequestID, ts int64, leaderIndex uint16) hashing.HashValue {
	var buf bytes.Buffer
	for i := range reqids {
		buf.Write(reqids[i].Bytes())
	}
	_ = util.WriteInt64(&buf, ts)
	_ = util.WriteUint16(&buf, leaderIndex)

	return *hashing.HashData(buf.Bytes())
}
