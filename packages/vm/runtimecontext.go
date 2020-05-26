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

// implements VMInputs interface
type RuntimeContext struct {
	// inputs
	ProgramHash   hashing.HashValue
	Address       address.Address
	Color         balance.Color
	Balances      map[valuetransaction.ID][]*balance.Balance
	RewardAddress address.Address
	Requests      []sctransaction.RequestRef
	Timestamp     time.Time
	VariableState state.VariableState // input/output
	OnFinish      func()
	Log           *logger.Logger

	// outputs
	ResultTransaction *sctransaction.Transaction
	StateUpdates      []state.StateUpdate
}

type Processor interface {
	Run(ctx *VMContext) state.StateUpdate
}

type VMContext struct {
	Address       address.Address
	Color         balance.Color
	Builder       *TransactionBuilder
	Timestamp     time.Time
	Request       sctransaction.RequestRef
	VariableState state.VariableState
	Log           *logger.Logger
}

func (ctx *RuntimeContext) BatchHash() *hashing.HashValue {
	var buf bytes.Buffer
	for _, reqRef := range ctx.Requests {
		buf.Write(reqRef.RequestId().Bytes())
	}
	_ = util.WriteUint64(&buf, uint64(ctx.Timestamp.UnixNano()))
	return hashing.HashData(buf.Bytes())
}
