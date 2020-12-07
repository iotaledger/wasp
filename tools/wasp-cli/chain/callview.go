package chain

import (
	"encoding/json"
	"os"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func callViewCmd(args []string) {
	if len(args) < 2 {
		log.Fatal("Usage: %s chain call-view <name> <funcname> [params]", os.Args[0])
	}
	r, err := SCClient(coretypes.Hn(args[0])).CallView(args[1], encodeParams(args[2:]))
	log.Check(err)
	log.Check(json.NewEncoder(os.Stdout).Encode(r))
}

func encodeParams(params []string) dict.Dict {
	d := dict.New()
	if len(params)%4 != 0 {
		log.Fatal("Params format: <type> <key> <type> <value> ...")
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
