package dwfcmd

import (
	"fmt"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"os"
	"strconv"
	"time"

	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfclient"
	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
)

func donateCmd(args []string) {
	if len(args) != 2 {
		fmt.Printf("Usage: %s dwf donate <amount> <feedback>\n", os.Args[0])
		os.Exit(1)
	}

	amount, err := strconv.Atoi(args[0])
	check(err)

	feedback := args[1]

	//util.WithSCRequest(dwf.Config, func() (*sctransaction.Transaction, error) {
	//	return dwf.Client().Donate(dwfclient.DonateParams{
	//		Amount:   int64(amount),
	//		Feedback: feedback,
	//	})
	//})
	tx, err := dwf.Client().Donate(dwfclient.DonateParams{
		Amount:            int64(amount),
		Feedback:          feedback,
		WaitForCompletion: true,
		PublisherHosts:    config.CommitteeNanomsg(dwf.Config.Committee()),
		PublisherQuorum:   2,
		Timeout:           30 * time.Second,
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("success. Request transaction id: %s\n", tx.ID().String())
}
