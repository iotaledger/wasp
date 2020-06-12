package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

//node messages represented as internal committee messages

type StateTransactionMsg struct {
	*sctransaction.Transaction
	Balances map[valuetransaction.ID][]*balance.Balance // may be nil
}

type BalancesMsg struct {
	Balances map[valuetransaction.ID][]*balance.Balance
}

type RequestMsg struct {
	*sctransaction.Transaction
	Index uint16
}

func (reqMsg *RequestMsg) RequestId() *sctransaction.RequestId {
	ret := sctransaction.NewRequestId(reqMsg.Transaction.ID(), reqMsg.Index)
	return &ret
}
