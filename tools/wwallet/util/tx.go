package util

import (
	"fmt"
	"os"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
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

	if config.WaitForConfirmation {
		check(config.GoshimmerClient().WaitForConfirmation(tx.ID()))
	}
}

func WithSCRequest(sc *config.SCConfig, f func() (*sctransaction.Transaction, error)) *sctransaction.Transaction {
	if config.WaitForConfirmation {
		tx, err := waspapi.RunAndWaitForRequestProcessedMulti(config.CommitteeNanomsg(sc.Committee()), sc.Address(), 0, 20*time.Second, f)
		check(err)
		return tx
	}
	tx, err := f()
	check(err)
	return tx
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
