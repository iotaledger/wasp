package util

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"os"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/mr-tron/base58"
)

func ValueFromString(vtype string, s string) []byte {
	switch vtype {
	case "color":
		col, err := ledgerstate.ColorFromBase58EncodedString(s)
		log.Check(err)
		return col.Bytes()
	case "agentid":
		agentid, err := coretypes.NewAgentIDFromString(s)
		log.Check(err)
		return agentid.Bytes()
	case "file":
		return ReadFile(s)
	case "string":
		return []byte(s)
	case "base58":
		b, err := base58.Decode(s)
		log.Check(err)
		return b
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
	case "string":
		return fmt.Sprintf("%q", string(v))
	}
	log.Fatal("ValueToString: No handler for type %s", vtype)
	return ""
}

func EncodeParams(params []string) dict.Dict {
	d := dict.New()
	if len(params)%4 != 0 {
		log.Fatal("Params format: <type> <key> <type> <value> ...")
	}
	for i := 0; i < len(params)/4; i++ {
		ktype := params[i*4]
		k := params[i*4+1]
		vtype := params[i*4+2]
		v := params[i*4+3]

		key := kv.Key(ValueFromString(ktype, k))
		val := ValueFromString(vtype, v)
		d.Set(key, val)
	}
	return d
}

func PrintDictAsJson(d dict.Dict) {
	log.Check(json.NewEncoder(os.Stdout).Encode(d))
}

func UnmarshalDict() dict.Dict {
	var d dict.Dict
	log.Check(json.NewDecoder(os.Stdin).Decode(&d))
	return d
}
