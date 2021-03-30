package util

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func PostTransaction(tx *ledgerstate.Transaction) {
	WithTransaction(func() (*ledgerstate.Transaction, error) {
		return tx, config.GoshimmerClient().PostTransaction(tx)
	})
}

func WithTransaction(f func() (*ledgerstate.Transaction, error)) *ledgerstate.Transaction {
	tx, err := f()
	log.Check(err)

	if config.WaitForCompletion {
		log.Check(config.GoshimmerClient().WaitForConfirmation(tx.ID()))
	}

	return tx
}

func WithSCTransaction(chainID coretypes.ChainID, f func() (*ledgerstate.Transaction, error), forceWait ...bool) *ledgerstate.Transaction {
	tx, err := f()
	log.Check(err)
	log.Printf("Posted transaction %s\n", tx.ID())

	if config.WaitForCompletion || len(forceWait) > 0 {
		log.Printf("Waiting for tx requests to be processed...\n")
		log.Check(config.WaspClient().WaitUntilAllRequestsProcessed(chainID, tx, 1*time.Minute))
	}

	return tx
}
