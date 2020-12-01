package util

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
)

func ValueFromString(vtype string, s string) []byte {
	switch vtype {
	case "color":
		col, err := util.ColorFromString(s)
		check(err)
		return col.Bytes()
	case "agentid":
		agentid, err := coretypes.AgentIDFromBase58(s)
		check(err)
		return agentid.Bytes()
	case "string":
		return []byte(s)
	}
	check(fmt.Errorf("ValueFromString: No handler for type %s", vtype))
	return nil
}

func ValueToString(vtype string, v []byte) string {
	switch vtype {
	case "color":
		col, _, err := balance.ColorFromBytes(v)
		check(err)
		return col.String()
	case "int":
		n, _ := util.Int64From8Bytes(v)
		return fmt.Sprintf("%d", n)
	}
	check(fmt.Errorf("ValueToString: No handler for type %s", vtype))
	return ""
}
