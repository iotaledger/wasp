package dwfcmd

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/tools/wwallet/sc/dwf"
	"github.com/iotaledger/wasp/tools/wwallet/util"
)

func statusCmd(args []string) {
	status, err := dwf.Client().FetchStatus()
	check(err)

	util.DumpSCStatus(dwf.Config, status.SCStatus)
	fmt.Printf("  amount of records: %d\n", status.NumRecords)
	fmt.Printf("  max donation: %d IOTAs\n", status.MaxDonation)
	fmt.Printf("  total donations: %d IOTAs\n", status.TotalDonations)
	fmt.Printf("  latest %d donations:\n", len(status.LastRecordsDesc))
	for _, di := range status.LastRecordsDesc {
		fmt.Printf("  - When: %s\n", di.When.UTC().Format(time.RFC3339))
		fmt.Printf("    Amount: %d IOTAs\n", di.Amount)
		fmt.Printf("    Sender: %s\n", di.Sender)
		fmt.Printf("    Feedback: %s\n", di.Feedback)
		if len(di.Error) > 0 {
			fmt.Printf("    Error: %s\n", di.Error)
		}
	}
}
