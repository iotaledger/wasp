package level1

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

// Level1Client is an interface to interact with the IOTA level 1 ledger
type Level1Client interface {
	// RequestFunds requests some funds from the testnet Faucet
	RequestFunds(targetAddress *address.Address) error

	// GetConfirmedAccountOutputs fetches all confirmed outputs belonging to the given address
	GetConfirmedAccountOutputs(address *address.Address) (map[transaction.OutputID][]*balance.Balance, error)

	// PostTransaction posts a transaction to the ledger
	PostTransaction(tx *transaction.Transaction) error

	// PostAndWaitForConfirmation posts a transaction to the ledger and blocks until it is confirmed
	PostAndWaitForConfirmation(tx *transaction.Transaction) error

	// WaitForConfirmation blocks until a transaction is confirmed in the ledger
	WaitForConfirmation(txid transaction.ID) error
}
