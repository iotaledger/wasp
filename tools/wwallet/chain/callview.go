package chain

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wwallet/util"
)

func callViewCmd(args []string) {
	if len(args) < 2 {
		check(fmt.Errorf("Usage: %s chain call-view <name> <funcname> [params]", os.Args[0]))
	}
	r, err := SCClient(coretypes.Hn(args[0])).CallView(args[1], encodeParams(args[2:]))
	check(err)
	check(json.NewEncoder(os.Stdout).Encode(r))
}

func encodeParams(params []string) dict.Dict {
	d := dict.New()
	if len(params)%4 != 0 {
		check(fmt.Errorf("Params format: <type> <key> <type> <value> ..."))
	}
	for i := 0; i < len(params)/4; i++ {
		ktype := params[i*3]
		k := params[i*3+1]
		vtype := params[i*3+2]
		v := params[i*3+3]

		key := kv.Key(util.ValueFromString(ktype, k))
		val := util.ValueFromString(vtype, v)
		d.Set(key, val)
	}
	return d
}
