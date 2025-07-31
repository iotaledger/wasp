// Package util provides utility functions and common helper methods
// used throughout the wasp-cli tool.
package util

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/kv/dict"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

//nolint:funlen,gocyclo
func ValueFromString(vtype, s string) []byte {
	switch strings.ToLower(vtype) {
	case "address":
		addr, err := cryptolib.NewAddressFromHexString(s)
		log.Check(err)

		return codec.Encode(addr)
	case "agentid":
		return codec.Encode(AgentIDFromString(s))
	case "bigint":
		n, ok := new(big.Int).SetString(s, 10)
		if !ok {
			log.Fatal("error converting to bigint")
		}

		return codec.Encode(n)
	case "bool":
		b, err := strconv.ParseBool(s)
		log.Check(err)
		return codec.Encode(b)
	case "bytes", "hex":
		b, err := cryptolib.DecodeHex(s)
		log.Check(err)
		return codec.Encode(b)
	case "chainid":
		chainid, err := isc.ChainIDFromString(s)
		log.Check(err)
		return codec.Encode(chainid)
	case "dict":
		d := dict.Dict{}
		err := d.UnmarshalJSON([]byte(s))
		log.Check(err)
		return codec.Encode(d)
	case "file":
		return ReadFile(s)
	case "hash":
		hash, err := hashing.HashValueFromHex(s)
		log.Check(err)
		return codec.Encode(hash)
	case "hname":
		hn, err := isc.HnameFromString(s)
		log.Check(err)
		return codec.Encode(hn)
	case "int8":
		n, err := strconv.ParseInt(s, 10, 8)
		log.Check(err)
		return codec.Encode[int8](int8(n))
	case "int16":
		n, err := strconv.ParseInt(s, 10, 16)
		log.Check(err)
		return codec.Encode[int16](int16(n))
	case "int32":
		n, err := strconv.ParseInt(s, 10, 32)
		log.Check(err)
		return codec.Encode[int32](int32(n))
	case "int64", "int":
		n, err := strconv.ParseInt(s, 10, 64)
		log.Check(err)
		return codec.Encode[int64](n)
	case "objectid":
		nidBytes, err := cryptolib.DecodeHex(s)
		log.Check(err)
		if len(nidBytes) != iotago.AddressLen {
			log.Fatal("invalid objectid length")
		}
		nid := [iotago.AddressLen]byte(nidBytes)
		return codec.Encode[iotago.ObjectID](nid)
	case "requestid":
		rid, err := isc.RequestIDFromString(s)
		log.Check(err)
		return codec.Encode(rid)
	case "string":
		return codec.Encode(s)
	case "tokenid":
		tidBytes, err := cryptolib.DecodeHex(s)
		log.Check(err)
		if len(tidBytes) != iotago.AddressLen {
			log.Fatal("invalid tokenid length")
		}

		tid := [iotago.AddressLen]byte(tidBytes)

		return codec.Encode(tid)
	case "uint8":
		n, err := strconv.ParseUint(s, 10, 8)
		log.Check(err)
		return codec.Encode[uint8](uint8(n))
	case "uint16":
		n, err := strconv.ParseUint(s, 10, 16)
		log.Check(err)
		return codec.Encode[uint16](uint16(n))
	case "uint32":
		n, err := strconv.ParseUint(s, 10, 32)
		log.Check(err)
		return codec.Encode[uint32](uint32(n))
	case "uint64":
		n, err := strconv.ParseUint(s, 10, 64)
		log.Check(err)
		return codec.Encode[uint64](n)
	}
	log.Fatalf("ValueFromString: No handler for type %s", vtype)
	return nil
}

//nolint:funlen,gocyclo
func ValueToString(vtype string, v []byte) string {
	switch strings.ToLower(vtype) {
	case "address":
		addr, err := codec.Decode[*cryptolib.Address](v)
		log.Check(err)
		return addr.String()
	case "coinbalances":
		cbs, err := codec.Decode[*isc.CoinBalances](v)
		log.Check(err)
		return cbs.String()
	case "assets":
		assets, err := codec.Decode[*isc.Assets](v)
		log.Check(err)
		return assets.String()
	case "agentid":
		aid, err := codec.Decode[isc.AgentID](v)
		log.Check(err)
		return aid.String()
	case "bigint":
		n, err := codec.Decode[*big.Int](v)
		log.Check(err)
		return n.String()
	case "bool":
		b, err := codec.Decode[bool](v)
		log.Check(err)
		if b {
			return "true"
		}
		return "false"
	case "bytes", "hex":
		b, err := codec.Decode[[]byte](v)
		log.Check(err)
		return cryptolib.EncodeHex(b)
	case "chainid":
		cid, err := codec.Decode[isc.ChainID](v)
		log.Check(err)
		return cid.String()
	case "dict":
		d, err := codec.Decode[dict.Dict](v)
		log.Check(err)
		s, err := d.MarshalJSON()
		log.Check(err)
		return string(s)
	case "hash":
		hash, err := codec.Decode[hashing.HashValue](v)
		log.Check(err)
		return hash.String()
	case "hname":
		hn, err := codec.Decode[isc.Hname](v)
		log.Check(err)
		return hn.String()
	case "int8":
		n, err := codec.Decode[int8](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int16":
		n, err := codec.Decode[int16](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int32":
		n, err := codec.Decode[int32](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "int64", "int":
		n, err := codec.Decode[int64](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "objectid":
		nid, err := codec.Decode[iotago.ObjectID](v)
		log.Check(err)
		return nid.String()
	case "requestid":
		rid, err := codec.Decode[isc.RequestID](v)
		log.Check(err)
		return rid.String()
	case "string":
		return fmt.Sprintf("%q", string(v))
	case "tokenid":
		tid, err := codec.Decode[coin.Type](v)
		log.Check(err)
		return tid.String()
	case "uint8":
		n, err := codec.Decode[uint8](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "uint16":
		n, err := codec.Decode[uint16](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "uint32":
		n, err := codec.Decode[uint32](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	case "uint64":
		n, err := codec.Decode[uint64](v)
		log.Check(err)
		return fmt.Sprintf("%d", n)
	}

	log.Fatalf("ValueToString: No handler for type %s", vtype)
	return ""
}

func EncodeParams(params []string) isc.CallArguments {
	if len(params)%2 != 0 {
		log.Fatal("Params format: <type> <value> ...")
	}

	encodedParams := make(isc.CallArguments, 0, len(params)/2)

	for i := 0; i < len(params)/2; i++ {
		vtype := params[i*2]
		v := params[i*2+1]

		val := ValueFromString(vtype, v)
		encodedParams = append(encodedParams, val)
	}

	return encodedParams
}

func PrintCallResultsAsJSON(res isc.CallResults) {
	log.Check(json.NewEncoder(os.Stdout).Encode(models.ToCallResultsJSON(res)))
}

func ReadCallResultsAsJSON() isc.CallArguments {
	var args models.CallResultsJSON
	log.Check(json.NewDecoder(os.Stdin).Decode(&args))
	return lo.Must(args.ToCallResults())
}

func AgentIDFromArgs(args []string) isc.AgentID {
	if len(args) == 0 {
		return isc.NewAddressAgentID(wallet.Load().Address())
	}
	return AgentIDFromString(args[0])
}

func AgentIDFromString(s string) isc.AgentID {
	if s == "common" {
		return accounts.CommonAccount()
	}

	agentID, err := isc.AgentIDFromString(s)
	log.Check(err, "cannot parse AgentID")
	return agentID
}
