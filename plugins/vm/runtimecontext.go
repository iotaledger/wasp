package vm

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

// implements VMInputs interface
type RuntimeContext struct {
	// inputs
	ProgramHash     hashing.HashValue
	Address         address.Address
	Color           balance.Color
	Balances        map[valuetransaction.ID][]*balance.Balance
	RewardAddress   address.Address
	RequestMessages []*committee.RequestMsg
	Timestamp       time.Time
	VariableState   state.VariableState
	OnFinish        func()
	Log             *logger.Logger

	// outputs, result of calculations
	ResultTransaction *sctransaction.Transaction
	StateUpdates      []state.StateUpdate
}

type Processor interface {
	Run(inputs *VMInput) *VMOutput
}

type VMInput struct {
	Address       address.Address
	Color         balance.Color
	Accounts      *accounts
	Timestamp     time.Time
	RequestMsg    *committee.RequestMsg
	VariableState state.VariableState
	Log           *logger.Logger
}

type VMOutput struct {
	Inputs       *VMInput
	StateUpdates []state.StateUpdate
}

func (ctx *RuntimeContext) BatchHash() hashing.HashValue {
	var buf bytes.Buffer
	for _, msg := range ctx.RequestMessages {
		buf.Write(msg.RequestId()[:])
	}
	_ := util.WriteUint64(&buf, uint64(ctx.Timestamp.UnixNano()))
	return *hashing.HashData(buf.Bytes())
}

func (ctx *RuntimeContext) taskName() string {
	return ctx.Address.String() + "." + ctx.BatchHash().String()
}
