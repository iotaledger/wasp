package metrics

import (
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var nodeconnMetricsCmd = &cobra.Command{
	Use:   "nodeconn",
	Short: "Show current value of collected metrics of connection to L1",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := config.WaspClient()
		if chainIDStr == "" {
			subscribed, msgMetrics, err := client.GetNodeConnectionMetrics()
			log.Check(err)
			log.Printf("Following chains subscribed to L1 events:\n")
			for _, s := range subscribed {
				log.Printf("\t%s\n", s)
			}
			printMessageMetrics(msgMetrics)
		} else {
			chid, err := iscp.ChainIDFromBase58(chainIDStr)
			log.Check(err)
			msgMetrics, err := client.GetChainNodeConnectionMetrics(chid)
			log.Check(err)
			printMessageMetrics(msgMetrics)
		}
	},
}

func printMessageMetrics(table [][]string) {
	header := []string{"Message name", "", "Total", "Last time", "Last message"}
	for i := range table {
		table[i][4] = strings.Replace(table[i][4], "\n", " ", -1)
	}
	log.PrintTable(header, table)
}
