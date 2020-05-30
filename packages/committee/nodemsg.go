package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type StateTransactionMsg struct {
	*sctransaction.Transaction
}

type BalancesMsg struct {
	Balances map[valuetransaction.ID][]*balance.Balance
}

type RequestMsg struct {
	*sctransaction.Transaction
	Index   uint16
	Outputs map[valuetransaction.ID][]*balance.Balance
}

func (reqMsg *RequestMsg) RequestId() *sctransaction.RequestId {
	ret := sctransaction.NewRequestId(reqMsg.Transaction.ID(), reqMsg.Index)
	return &ret
}
