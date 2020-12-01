package util

import (
	"fmt"
	"os"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/tools/wwallet/config"
)

func PostTransaction(tx *transaction.Transaction) {
	WithTransaction(func() (*transaction.Transaction, error) {
		return tx, config.GoshimmerClient().PostTransaction(tx)
	})
}

func WithTransaction(f func() (*transaction.Transaction, error)) {
	tx, err := f()
	check(err)

	if config.WaitForCompletion {
		check(config.GoshimmerClient().WaitForConfirmation(tx.ID()))
	}
}

func WithSCTransaction(f func() (*sctransaction.Transaction, error)) {
	tx, err := f()
	check(err)

	if config.WaitForCompletion {
		check(config.WaspClient().WaitUntilAllRequestsProcessed(tx, 1*time.Minute))
	}
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
