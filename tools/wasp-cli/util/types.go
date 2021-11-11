package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/mr-tron/base58"
)

func ValueFromString(vtype, s string) []byte {
	switch vtype {
	case "uint64":
		n, err := strconv.Atoi(s)
		log.Check(err)
		return codec.EncodeUint64(uint64(n))
	case "uint32":
		n, err := strconv.Atoi(s)
		log.Check(err)
		return codec.EncodeUint32(uint32(n))
	case "int32":
		n, err := strconv.Atoi(s) //nolint:gosec // potential int32 overflow
		log.Check(err)
		return codec.EncodeInt32(int32(n))
	case "int", "int64":
		n, err := strconv.Atoi(s)
		log.Check(err)
		return codec.EncodeInt64(int64(n))
	case "bool":
		b, err := strconv.ParseBool(s)
		log.Check(err)
		return codec.EncodeBool(b)
	case "color":
		col, err := ledgerstate.ColorFromBase58EncodedString(s)
		log.Check(err)
		return col.Bytes()
	case "agentid":
		agentid, err := iscp.NewAgentIDFromString(s)
		log.Check(err)
		return agentid.Bytes()
	case "file":
		return ReadFile(s)
	case "string":
		return []byte(s)
	case "bytes", "base58":
		b, err := base58.Decode(s)
		log.Check(err)
		return b
	}
	log.Fatalf("ValueFromString: No handler for type %s", vtype)
	return nil
}

func ValueToString(vtype string, v []byte) string {
	switch vtype {
	case "color":
		col, err := codec.DecodeColor(v)
		log.Check(err)
		return col.String()
	case "agentid":
		aid, err := codec.DecodeAgentID(v)
		log.Check(err)
		return aid.String()
	case "uint64":
		n, err := codec.DecodeUint64(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int16":
		n, err := codec.DecodeInt16(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int", "int64":
		n, err := codec.DecodeInt64(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "string":
		return fmt.Sprintf("%q", string(v))
	case "bytes", "base58":
		return base58.Encode(v)
	}
	log.Fatalf("ValueToString: No handler for type %s", vtype)
	return ""
}

func EncodeParams(params []string) dict.Dict {
	d := dict.New()
	if len(params)%4 != 0 {
		log.Fatalf("Params format: <type> <key> <type> <value> ...")
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

func PrintDictAsJSON(d dict.Dict) {
	log.Check(json.NewEncoder(os.Stdout).Encode(d))
}

func UnmarshalDict() dict.Dict {
	var d dict.Dict
	log.Check(json.NewDecoder(os.Stdin).Decode(&d))
	return d
}
