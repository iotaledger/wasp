package chain

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func listCmd(args []string) {
	client := config.WaspClient()
	chains, err := client.GetChainRecordList()
	log.Check(err)
	log.Printf("Total %d chains in wasp node %s\n", len(chains), client.BaseURL())
	showChainInfo(chains)
}
