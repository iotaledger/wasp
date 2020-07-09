package client

import (
	"fmt"
	"os"
	"strconv"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/util"
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

	util.PostTransaction(&waspapi.RequestBlockJson{
		Address:     config.GetSCAddress().String(),
		RequestCode: fairroulette.RequestPlaceBet,
		AmountIotas: int64(amount),
		Vars: map[string]interface{}{
			fairroulette.ReqVarColor: int64(color),
		},
	})
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
