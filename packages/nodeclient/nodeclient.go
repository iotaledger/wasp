package nodeclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

type NodeClient interface {
	RequestFunds(targetAddress *address.Address) error
	GetAccountOutputs(address *address.Address) (map[transaction.OutputID][]*balance.Balance, error)
	PostTransaction(tx *transaction.Transaction) error
	PostAndWaitForConfirmation(tx *transaction.Transaction) error
}

