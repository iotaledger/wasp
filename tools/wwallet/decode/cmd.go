package decode

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wwallet/util"
	"github.com/spf13/pflag"
)

func InitCommands(commands map[string]func([]string), flags *pflag.FlagSet) {
	commands["decode"] = decodeCmd
}

func decodeCmd(args []string) {
	var d dict.Dict
	check(json.NewDecoder(os.Stdin).Decode(&d))

	c := codec.NewMustCodec(d)

	if len(args) == 2 {
		ktype := args[0]
		vtype := args[1]

		c.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
			skey := util.ValueToString(ktype, []byte(key))
			sval := util.ValueToString(vtype, value)
			fmt.Printf("%s: %s\n", skey, sval)
			return true
		})
		return
	}

	if len(args) < 3 || len(args)%3 != 0 {
		usage()
	}

	for i := 0; i < len(args)/2; i++ {
		ktype := args[i*2]
		skey := args[i*2+1]
		vtype := args[i*2+2]

		key := kv.Key(util.ValueFromString(ktype, skey))
		val := c.Get(key)
		if val == nil {
			fmt.Printf("%s: <nil>\n", skey)
		} else {
			fmt.Printf("%s: %s\n", skey, util.ValueToString(vtype, val))
		}
	}
}

func usage() {
	fmt.Printf("Usage: %s decode <type> <key> <type> [...]\n", os.Args[0])
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
