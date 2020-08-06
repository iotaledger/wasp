package util

import (
	"fmt"
	"os"

	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
)

func PostTransaction(req *waspapi.RequestBlockJson) {
	tx, err := waspapi.CreateRequestTransaction(
		config.GoshimmerClient(),
		wallet.Load().SignatureScheme(),
		[]*waspapi.RequestBlockJson{req},
	)
	check(err)

	check(nodeapi.PostTransaction(config.GoshimmerApi(), tx.Transaction))
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
