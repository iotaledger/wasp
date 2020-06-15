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
)

// task context (for batch of requests)
type VMTask struct {
	// inputs (immutable)
	LeaderPeerIndex uint16
	ProgramHash     hashing.HashValue
	Address         address.Address
	Color           balance.Color
	Balances        map[valuetransaction.ID][]*balance.Balance
	OwnerAddress    address.Address
	RewardAddress   address.Address
	Requests        []sctransaction.RequestRef
	Timestamp       int64
	VariableState   state.VariableState // input immutable
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

// context of one VM call (for one request)
type VMContext struct {
	// invariant through the batch
	// address of the smart contract
	Address address.Address
	// tx builder to build the final transaction
	TxBuilder *TransactionBuilder
	// timestamp of the batch
	Timestamp int64
	// initial state of the batch
	VariableState state.VariableState
	// log
	Log *logger.Logger
	// set for each call
	Request sctransaction.RequestRef
	// IsEmpty state update upon call, result of the call.
	StateUpdate state.StateUpdate
}

func BatchHash(reqids []sctransaction.RequestId, ts int64) hashing.HashValue {
	var buf bytes.Buffer
	for i := range reqids {
		buf.Write(reqids[i].Bytes())
	}
	_ = util.WriteUint64(&buf, uint64(ts))
	return *hashing.HashData(buf.Bytes())
}
