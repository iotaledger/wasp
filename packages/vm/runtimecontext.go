package vm

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

type VMTask struct {
	// inputs
	LeaderPeerIndex uint16
	ProgramHash     hashing.HashValue
	Address         address.Address
	Color           balance.Color
	Balances        map[valuetransaction.ID][]*balance.Balance
	RewardAddress   address.Address
	Requests        []sctransaction.RequestRef
	Timestamp       time.Time
	VariableState   state.VariableState // input
	Log             *logger.Logger
	// call when finished
	OnFinish func()
	// outputs
	ResultTransaction *sctransaction.Transaction
	ResultBatch       state.Batch
}

type Processor interface {
	Run(ctx *VMContext)
}

// context of one VM call
type VMContext struct {
	// same context duting batch
	Address       address.Address
	Color         balance.Color
	TxBuilder     *TransactionBuilder
	Timestamp     time.Time
	VariableState state.VariableState
	Log           *logger.Logger
	// input for one call
	Request sctransaction.RequestRef
	// output of one call
	StateUpdate state.StateUpdate
}

func BatchHash(reqids []sctransaction.RequestId, timestamp time.Time) hashing.HashValue {
	var buf bytes.Buffer
	for i := range reqids {
		buf.Write(reqids[i].Bytes())
	}
	_ = util.WriteUint64(&buf, uint64(timestamp.UnixNano()))
	return *hashing.HashData(buf.Bytes())
}
