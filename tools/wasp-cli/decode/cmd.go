package decode

import (
	"encoding/json"
	"os"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["decode"] = decodeCmd
}

func decodeCmd(args []string) {
	var d dict.Dict
	log.Check(json.NewDecoder(os.Stdin).Decode(&d))

	if len(args) == 2 {
		ktype := args[0]
		vtype := args[1]

		for key, value := range d {
			skey := util.ValueToString(ktype, []byte(key))
			sval := util.ValueToString(vtype, value)
			log.Printf("%s: %s\n", skey, sval)
		}
		return
	}

	if len(args) < 3 || len(args)%3 != 0 {
		log.Usage("%s decode <type> <key> <type> [...]\n", os.Args[0])
	}

	for i := 0; i < len(args)/2; i++ {
		ktype := args[i*2]
		skey := args[i*2+1]
		vtype := args[i*2+2]

		key := kv.Key(util.ValueFromString(ktype, skey))
		val := d.MustGet(key)
		if val == nil {
			log.Printf("%s: <nil>\n", skey)
		} else {
			log.Printf("%s: %s\n", skey, util.ValueToString(vtype, val))
		}
	}
}
