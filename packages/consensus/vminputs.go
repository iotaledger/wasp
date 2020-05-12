package consensus

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

// implements VMInputs interface
type runtimeContext struct {
	// address of the smart contract
	address *address.Address
	// color of the smart contract
	color *balance.Color
	// outputs/balances of the account address
	balances map[valuetransaction.ID][]*balance.Balance
	// reward address
	rewardAddress *address.Address
	// input. Leader where results must be sent
	leaderPeerIndex uint16
	// input. Requests, may be a batch
	reqMsg []*committee.RequestMsg
	// timestamp of the context. Imposed by the leader
	timestamp time.Time
	// current state represented by the stateTx and variableState
	stateTx       *sctransaction.Transaction
	variableState state.VariableState
	// output of the computation, represented by the resultTx and stateUpdate
	log *logger.Logger
}

func (ctx *runtimeContext) Address() *address.Address {
	return ctx.address
}

func (ctx *runtimeContext) Color() *balance.Color {
	return ctx.color
}

func (ctx *runtimeContext) Balances() map[valuetransaction.ID][]*balance.Balance {
	return ctx.balances
}

func (ctx *runtimeContext) RewardAddress() *address.Address {
	return ctx.rewardAddress
}

func (ctx *runtimeContext) RequestMsg() []*committee.RequestMsg {
	return ctx.reqMsg
}

func (ctx *runtimeContext) Timestamp() time.Time {
	return ctx.timestamp
}

func (ctx *runtimeContext) StateTransaction() *sctransaction.Transaction {
	return ctx.stateTx
}

func (ctx *runtimeContext) VariableState() state.VariableState {
	return ctx.variableState
}

func (ctx *runtimeContext) Log() *logger.Logger {
	return ctx.log
}
