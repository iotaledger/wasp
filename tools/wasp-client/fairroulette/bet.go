package fairroulette

import (
	"fmt"
	"os"
	"strconv"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/util"
)

func betCmd(args []string) {
	if len(args) != 2 {
		scConfig.PrintUsage("bet <color> <amount>")
		os.Exit(1)
	}

	color, err := strconv.Atoi(args[0])
	check(err)
	amount, err := strconv.Atoi(args[1])
	check(err)

	util.PostRequest(&waspapi.RequestBlockJson{
		Address:     config.GetFRAddress().String(),
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
