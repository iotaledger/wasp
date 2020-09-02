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

func donateCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s dwf donate <amount> <feedback>\n", os.Args[0])
		os.Exit(1)
	}

	amount, err := strconv.Atoi(args[0])
	check(err)

	feedback := args[1]

	util.WithSCRequest(dwf.Config, func() (*sctransaction.Transaction, error) {
		return dwf.Client().Donate(dwfclient.DonateParams{
			Amount:   int64(amount),
			Feedback: feedback,
		})
	})
}
