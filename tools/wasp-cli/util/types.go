package util

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func ValueFromString(vtype string, s string) []byte {
	switch vtype {
	case "color":
		col, err := util.ColorFromString(s)
		log.Check(err)
		return col.Bytes()
	case "agentid":
		agentid, err := coretypes.NewAgentIDFromString(s)
		log.Check(err)
		return agentid.Bytes()
	case "string":
		return []byte(s)
	}
	log.Fatal("ValueFromString: No handler for type %s", vtype)
	return nil
}

func ValueToString(vtype string, v []byte) string {
	switch vtype {
	case "color":
		col, _, err := balance.ColorFromBytes(v)
		log.Check(err)
		return col.String()
	case "int":
		n, _ := util.Int64From8Bytes(v)
		return fmt.Sprintf("%d", n)
	}
	log.Fatal("ValueToString: No handler for type %s", vtype)
	return ""
}
