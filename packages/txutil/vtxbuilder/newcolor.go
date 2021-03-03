package vtxbuilder

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/client/level1"
)

// NewColoredTokensTransaction mints specified amount of colored tokens
// from iotas in the address corresponding to sigScheme.
// It returns a value transaction with empty data payload (not sc transaction)
func NewColoredTokensTransaction(client level1.Level1Client, sigScheme signaturescheme.SignatureScheme, amount int64) (*valuetransaction.Transaction, error) {
	addr := sigScheme.Address()
	allOuts, err := client.GetConfirmedAccountOutputs(&addr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}
	txb, err := NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}
	if err := txb.MintColoredTokens(addr, balance.ColorIOTA, amount); err != nil {
		return nil, err
	}
	tx := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(sigScheme)
	return tx, nil
}
