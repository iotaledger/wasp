package util

import (
	"fmt"
	"os"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
)

func PostRequest(req *waspapi.RequestBlockJson) {
	tx, err := waspapi.CreateRequestTransaction(
		config.GoshimmerClient(),
		wallet.Load().SignatureScheme(),
		[]*waspapi.RequestBlockJson{req},
	)
	check(err)

	check(config.GoshimmerClient().PostTransaction(tx.Transaction))
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
