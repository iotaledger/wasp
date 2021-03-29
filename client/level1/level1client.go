package level1

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/plugins/webapi/value"
)

// Level1Client is an interface to interact with the IOTA level 1 ledger
type Level1Client interface {
	// RequestFunds requests some funds from the testnet Faucet
	RequestFunds(targetAddress ledgerstate.Address) error

	// GetConfirmedAccountOutputs fetches all confirmed outputs belonging to the given address
	GetConfirmedOutputs(address ledgerstate.Address) ([]value.Output, error)

	// PostTransaction posts a transaction to the ledger
	PostTransaction(tx *ledgerstate.Transaction) error

	// PostAndWaitForConfirmation posts a transaction to the ledger and blocks until it is confirmed
	PostAndWaitForConfirmation(tx *ledgerstate.Transaction) error

	// WaitForConfirmation blocks until a transaction is confirmed in the ledger
	WaitForConfirmation(txid ledgerstate.TransactionID) error
}
