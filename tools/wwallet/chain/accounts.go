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
)

func listAccountsCmd(args []string) {
	ret, err := SCClient(accountsc.Interface.Hname()).CallView(accountsc.FuncAccounts, nil)
	check(err)
	codec.NewMustCodec(ret).Iterate(kv.EmptyPrefix, func(k kv.Key, v []byte) bool {
		agentId, err := coretypes.NewAgentIDFromBytes([]byte(k))
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%s (%s)\n", agentId.Base58(), agentId.String())
		return true
	})
}

func balanceCmd(args []string) {
	if len(args) != 1 {
		fmt.Printf("Usage: %s chain balance <agentid>\n", os.Args[0])
		os.Exit(1)
	}

	agentID := parseAgentID(args[0])

	ret, err := SCClient(accountsc.Interface.Hname()).CallView(accountsc.FuncBalance, dict.FromGoMap(map[kv.Key][]byte{
		accountsc.ParamAgentID: agentID.Bytes(),
	}))
	check(err)
	codec.NewMustCodec(ret).Iterate(kv.EmptyPrefix, func(k kv.Key, v []byte) bool {
		color, _, err := balance.ColorFromBytes([]byte(k))
		check(err)
		bal, err := util.Uint64From8Bytes(v)
		check(err)
		fmt.Printf("%s: %d\n", color, bal)
		return true
	})
}

func parseAgentID(s string) coretypes.AgentID {
	agentid, err := coretypes.AgentIDFromBase58(s)
	check(err)
	return agentid
}
