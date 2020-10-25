package chain

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type StateTransactionMsg struct {
	*sctransaction.Transaction
}

type TransactionInclusionLevelMsg struct {
	TxId  *valuetransaction.ID
	Level byte
}

type BalancesMsg struct {
	Balances map[valuetransaction.ID][]*balance.Balance
}

type RequestMsg struct {
	*sctransaction.Transaction
	Index coretypes.Uint16
}

func (reqMsg *RequestMsg) RequestId() *coretypes.RequestID {
	ret := coretypes.NewRequestID(reqMsg.Transaction.ID(), reqMsg.Index)
	return &ret
}

func (reqMsg *RequestMsg) RequestBlock() *sctransaction.RequestBlock {
	return reqMsg.Requests()[reqMsg.Index]
}

func (reqMsg *RequestMsg) Timelock() uint32 {
	return reqMsg.RequestBlock().Timelock()
}
