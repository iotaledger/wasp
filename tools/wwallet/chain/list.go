package chain

import (
	"github.com/iotaledger/wasp/tools/wwallet/config"
)

func listCmd(args []string) {
	chains, err := config.WaspClient().GetChainRecordList()
	check(err)
	for _, chain := range chains {
		showChainInfo(chain)
	}
}
