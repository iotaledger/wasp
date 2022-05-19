package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/mr-tron/base58"
)

//nolint:funlen
func ValueFromString(vtype, s string) []byte {
	switch vtype {
	case "address":
		prefix, addr, err := iotago.ParseBech32(s)
		log.Check(err)
		if parameters.L1 == nil {
			config.L1Client() // this will fill parameters.L1 with data from the L1 node
		}
		l1Prefix := parameters.L1.Protocol.Bech32HRP
		if prefix != l1Prefix {
			log.Fatalf("address prefix %s does not match L1 prefix %s", prefix, l1Prefix)
		}
		return iscp.BytesFromAddress(addr)
	case "agentid":
		agentid, err := iscp.NewAgentIDFromString(s)
		log.Check(err)
		return agentid.Bytes()
	case "bool":
		b, err := strconv.ParseBool(s)
		log.Check(err)
		return codec.EncodeBool(b)
	case "bytes", "base58":
		b, err := base58.Decode(s)
		log.Check(err)
		return b
	case "chainid":
		_, chainid, err := iotago.ParseBech32(s)
		log.Check(err)
		return iscp.BytesFromAddress(chainid)
	case "file":
		return ReadFile(s)
	case "hash":
		hash, err := hashing.HashValueFromHex(s)
		log.Check(err)
		return hash.Bytes()
	case "hname":
		hn, err := iscp.HnameFromString(s)
		log.Check(err)
		return hn.Bytes()
	case "int8":
		n, err := strconv.ParseInt(s, 10, 8)
		log.Check(err)
		return codec.EncodeInt8(int8(n))
	case "int16":
		n, err := strconv.ParseInt(s, 10, 16)
		log.Check(err)
		return codec.EncodeInt16(int16(n))
	case "int32":
		n, err := strconv.ParseInt(s, 10, 32)
		log.Check(err)
		return codec.EncodeInt32(int32(n))
	case "int64", "int":
		n, err := strconv.ParseInt(s, 10, 64)
		log.Check(err)
		return codec.EncodeInt64(n)
	case "requestid":
		rid, err := iscp.RequestIDFromString(s)
		log.Check(err)
		return rid.Bytes()
	case "string":
		return []byte(s)
	case "uint8":
		n, err := strconv.ParseUint(s, 10, 8)
		log.Check(err)
		return codec.EncodeUint8(uint8(n))
	case "uint16":
		n, err := strconv.ParseUint(s, 10, 16)
		log.Check(err)
		return codec.EncodeUint16(uint16(n))
	case "uint32":
		n, err := strconv.ParseUint(s, 10, 32)
		log.Check(err)
		return codec.EncodeUint32(uint32(n))
	case "uint64":
		n, err := strconv.ParseUint(s, 10, 64)
		log.Check(err)
		return codec.EncodeUint64(n)
	}
	log.Fatalf("ValueFromString: No handler for type %s", vtype)
	return nil
}

//nolint:funlen
func ValueToString(vtype string, v []byte) string {
	switch vtype {
	case "address":
		addr, err := codec.DecodeAddress(v)
		log.Check(err)
		if parameters.L1 == nil {
			config.L1Client() // this will fill parameters.L1 with data from the L1 node
		}
		return addr.Bech32(parameters.L1.Protocol.Bech32HRP)
	case "agentid":
		aid, err := codec.DecodeAgentID(v)
		log.Check(err)
		return aid.String()
	case "bool":
		b, err := codec.DecodeBool(v)
		log.Check(err)
		if b {
			return "true"
		}
		return "false"
	case "bytes", "base58":
		return base58.Encode(v)
	case "chainid":
		cid, err := codec.DecodeChainID(v)
		log.Check(err)
		return cid.String()
	case "hash":
		hash, err := codec.DecodeHashValue(v)
		log.Check(err)
		return hash.String()
	case "hname":
		hn, err := codec.DecodeHname(v)
		log.Check(err)
		return hn.String()
	case "int8":
		n, err := codec.DecodeInt8(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int16":
		n, err := codec.DecodeInt16(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int32":
		n, err := codec.DecodeInt32(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int64", "int":
		n, err := codec.DecodeInt64(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "requestid":
		rid, err := codec.DecodeRequestID(v)
		log.Check(err)
		return rid.String()
	case "string":
		return fmt.Sprintf("%q", string(v))
	case "uint8":
		n, err := codec.DecodeUint8(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "uint16":
		n, err := codec.DecodeUint16(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "uint32":
		n, err := codec.DecodeUint32(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "uint64":
		n, err := codec.DecodeUint64(v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
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
