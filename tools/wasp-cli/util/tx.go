package util

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction_old"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func PostTransaction(tx *transaction.Transaction) {
	WithTransaction(func() (*transaction.Transaction, error) {
		return tx, config.GoshimmerClient().PostTransaction(tx)
	})
}

func WithTransaction(f func() (*transaction.Transaction, error)) *transaction.Transaction {
	tx, err := f()
	log.Check(err)

	if config.WaitForCompletion {
		log.Check(config.GoshimmerClient().WaitForConfirmation(tx.ID()))
	}

	return tx
}

func WithSCTransaction(f func() (*sctransaction_old.TransactionEssence, error), forceWait ...bool) *sctransaction_old.TransactionEssence {
	tx, err := f()
	log.Check(err)

	log.Printf("Posted transaction %s\n", tx.ID())
	if config.WaitForCompletion || (len(forceWait) > 0) {
		log.Printf("Waiting for tx requests to be processed...\n")
		log.Check(config.WaspClient().WaitUntilAllRequestsProcessed(tx, 1*time.Minute))
	}

	return tx
}
