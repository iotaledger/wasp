package util

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
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

func WithSCTransaction(f func() (*sctransaction.Transaction, error)) *sctransaction.Transaction {
	tx, err := f()
	log.Check(err)

	if config.WaitForCompletion {
		log.Check(config.WaspClient().WaitUntilAllRequestsProcessed(tx, 1*time.Minute))
	}

	return tx
}
