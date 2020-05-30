package consensus

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

// balances must contain smart contract token
func (op *operator) validToken(balances map[valuetransaction.ID][]*balance.Balance) bool {
	found := false
	for _, bals := range balances {
		for _, b := range bals {
			if b.Color() == op.committee.Color() {
				if found || b.Value() != 1 {
					op.log.Errorf("expected exactly one token with color %s", op.committee.Color().String())
					return false
				}
				found = true
			}
		}
	}
	if !found {
		op.log.Errorf("token with color %s wasn't found", op.committee.Color().String())
		return false
	}
	return true
}
