package vtxbuilder

import (
	"fmt"

	"github.com/iotaledger/wasp/client/level1"
)

// NewColoredTokensTransaction mints specified amount of colored tokens
// from iotas in the address corresponding to sigScheme.
// It returns a value transaction with empty data payload (not sc transaction)
func NewColoredTokensTransaction(client level1.Level1Client, sigScheme signaturescheme.SignatureScheme, amount int64) (*valuetransaction.Transaction, error) {
	addr := sigScheme.Address()
	allOuts, err := client.GetConfirmedOutputs(&addr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}
	txb, err := NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}
	if err := txb.MintColoredTokens(addr, ledgerstate.ColorIOTA, amount); err != nil {
		return nil, err
	}
	tx := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(sigScheme)
	return tx, nil
}
