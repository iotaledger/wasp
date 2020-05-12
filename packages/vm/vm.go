package vm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"time"
)

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
	// result state update (without transaction hash yet)
	// it is a accumulated update of the batch of requests
	StateUpdate state.StateUpdate
}
