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
	// hash of the program to run
	ProgramHash hashing.HashValue
	// address of the smart contract
	Address address.Address
	// color of the smart contract
	Color balance.Color
	// outputs/balances of the account address
	Balances map[valuetransaction.ID][]*balance.Balance
	// reward address
	RewardAddress address.Address
	// input. Leader where results must be sent
	LeaderPeerIndex uint16
	// Batch of requests
	RequestMessages []*committee.RequestMsg
	// timestamp of the context. Imposed by the leader
	Timestamp time.Time
	// current state represented by the stateTx and variableState
	StateTransaction *sctransaction.Transaction
	VariableState    state.VariableState
	// output of the computation, represented by the resultTx and stateUpdate
	Log *logger.Logger

	// outputs, result of calculations
	ResultTransaction *sctransaction.Transaction
	// state update corresponding to requests
	StateUpdates []state.StateUpdate
}

type Processor interface {
	Run(inputs VMInputs) VMOutput
}

type VMInputs interface {
	// account address
	Address() *address.Address
	// color of the smart contracts
	Color() *balance.Color
	// balances/outputs of the account address (scid.Address())
	// imposed by the leader
	Balances() map[valuetransaction.ID][]*balance.Balance
	// reward address or nil of no rewards collected
	RewardAddress() *address.Address
	// timestamp imposed by the leader
	Timestamp() time.Time
	// batch of requests to run sequentially. .
	RequestMsg() []*committee.RequestMsg
	// the context state transaction
	StateTransaction() *sctransaction.Transaction
	// the context variable state
	VariableState() state.VariableState
	// log for VM
	Log() *logger.Logger
}

type VMOutput struct {
	// references to inouts
	Inputs VMInputs
	// result transaction (not signed)
	// accumulated final result of batch processing. It means the result transaction as inputs
	// has all outputs to the SC account address from all request
	// similarly outputs are consolidated, for example it should contain the only output of the SC token
	ResultTransaction *sctransaction.Transaction
	// state update corresponding to requests
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
