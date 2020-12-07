package chain

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func listAccountsCmd(args []string) {
	ret, err := SCClient(accountsc.Interface.Hname()).CallView(accountsc.FuncAccounts, nil)
	log.Check(err)

	log.Printf("Total %d accounts in chain %s\n", len(ret), GetCurrentChainID())

	header := []string{"agentid (b58)", "agentid (string)"}
	rows := make([][]string, len(ret))
	i := 0
	for k := range ret {
		agentId, err := coretypes.NewAgentIDFromBytes([]byte(k))
		if err != nil {
			panic(err.Error())
		}
		rows[i] = []string{agentId.Base58(), agentId.String()}
		i++
	}
	log.PrintTable(header, rows)
}

func balanceCmd(args []string) {
	if len(args) != 1 {
		log.Usage("%s chain balance <agentid>\n", os.Args[0])
	}

	agentID := parseAgentID(args[0])

	ret, err := SCClient(accountsc.Interface.Hname()).CallView(accountsc.FuncBalance, dict.FromGoMap(map[kv.Key][]byte{
		accountsc.ParamAgentID: agentID.Bytes(),
	}))
	log.Check(err)

	header := []string{"color", "amount"}
	rows := make([][]string, len(ret))
	i := 0
	for k, v := range ret {
		color, _, err := balance.ColorFromBytes([]byte(k))
		log.Check(err)
		bal, err := util.Uint64From8Bytes(v)
		log.Check(err)

		rows[i] = []string{color.String(), fmt.Sprintf("%d", bal)}
		i++
	}
	log.PrintTable(header, rows)
}

func parseAgentID(s string) coretypes.AgentID {
	agentid, err := coretypes.AgentIDFromBase58(s)
	log.Check(err)
	return agentid
}
