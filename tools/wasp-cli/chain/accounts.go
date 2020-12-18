package chain

import (
	"fmt"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func listAccountsCmd(args []string) {
	ret, err := SCClient(accountsc.Interface.Hname()).CallView(accountsc.FuncAccounts, nil)
	log.Check(err)

	log.Printf("Total %d account(s) in chain %s\n", len(ret), GetCurrentChainID())

	header := []string{"agentid"}
	rows := make([][]string, len(ret))
	i := 0
	for k := range ret {
		agentId, _, err := codec.DecodeAgentID([]byte(k))
		log.Check(err)
		rows[i] = []string{agentId.String()}
		i++
	}
	log.PrintTable(header, rows)
}

func balanceCmd(args []string) {
	if len(args) != 1 {
		log.Usage("%s chain balance <agentid>\n", os.Args[0])
	}

	agentID, err := coretypes.NewAgentIDFromString(args[0])
	log.Check(err)

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
