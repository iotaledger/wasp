package client

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
)

func BetCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s bet <color> <amount>\n", os.Args[0])
		os.Exit(1)
	}

	color, err := strconv.Atoi(args[0])
	check(err)
	amount, err := strconv.Atoi(args[1])
	check(err)

	check(placeBet(config.GoshimmerApi(), config.GetSCAddress(), color, amount, wallet.Load().SignatureScheme()))
}

func placeBet(goshimmerApi string, scAddress address.Address, color int, amount int, sigScheme signaturescheme.SignatureScheme) error {
	req := &waspapi.RequestBlockJson{
		Address:     scAddress.String(),
		RequestCode: fairroulette.RequestPlaceBet,
		AmountIotas: int64(amount),
		Vars: map[string]interface{}{
			fairroulette.ReqVarColor: int64(color),
		},
	}

	tx, err := waspapi.CreateRequestTransaction(goshimmerApi, sigScheme, []*waspapi.RequestBlockJson{req})
	if err != nil {
		return err
	}

	return nodeapi.PostTransaction(goshimmerApi, tx.Transaction)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
