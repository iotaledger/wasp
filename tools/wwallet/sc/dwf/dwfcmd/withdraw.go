package dwfcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfclient"
	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
	"github.com/iotaledger/wasp/tools/wwallet/util"
)

func withdrawCmd(args []string) {
	if len(args) != 1 {
		fmt.Printf("Usage: %s dwf withdraw <amount>\n", os.Args[0])
		os.Exit(1)
	}

	amount, err := strconv.Atoi(args[0])
	check(err)

	util.WithSCRequest(dwf.Config, func() (*sctransaction.Transaction, error) {
		return dwf.Client().Withdraw(dwfclient.WithdrawParams{
			Amount: int64(amount),
		})
	})
}
