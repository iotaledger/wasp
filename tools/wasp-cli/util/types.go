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
)

//nolint:funlen,gocyclo
func ValueFromString(vtype, s string) ([]byte, error) {
	switch strings.ToLower(vtype) {
	case "address":
		addr, err := cryptolib.NewAddressFromHexString(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(addr), nil
	case "agentid":
		agentID, err := AgentIDFromString(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(agentID), nil
	case "bigint":
		n, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("error converting to bigint")
		}
		return codec.Encode(n), nil
	case "bool":
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(b), nil
	case "bytes", "hex":
		b, err := cryptolib.DecodeHex(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(b), nil
	case "chainid":
		chainid, err := isc.ChainIDFromString(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(chainid), nil
	case "dict":
		d := dict.Dict{}
		err := d.UnmarshalJSON([]byte(s))
		if err != nil {
			return nil, err
		}
		return codec.Encode(d), nil
	case "file":
		return ReadFile(s)
	case "hash":
		hash, err := hashing.HashValueFromHex(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(hash), nil
	case "hname":
		hn, err := isc.HnameFromString(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(hn), nil
	case "int8":
		n, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return nil, err
		}
		return codec.Encode[int8](int8(n)), nil
	case "int16":
		n, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return nil, err
		}
		return codec.Encode[int16](int16(n)), nil
	case "int32":
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, err
		}
		return codec.Encode[int32](int32(n)), nil
	case "int64", "int":
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return codec.Encode[int64](n), nil
	case "objectid":
		nidBytes, err := cryptolib.DecodeHex(s)
		if err != nil {
			return nil, err
		}
		if len(nidBytes) != iotago.AddressLen {
			return nil, fmt.Errorf("invalid objectid length")
		}
		nid := [iotago.AddressLen]byte(nidBytes)
		return codec.Encode[iotago.ObjectID](nid), nil
	case "requestid":
		rid, err := isc.RequestIDFromString(s)
		if err != nil {
			return nil, err
		}
		return codec.Encode(rid), nil
	case "string":
		return codec.Encode(s), nil
	case "tokenid":
		tidBytes, err := cryptolib.DecodeHex(s)
		if err != nil {
			return nil, err
		}
		if len(tidBytes) != iotago.AddressLen {
			return nil, fmt.Errorf("invalid tokenid length")
		}
		tid := [iotago.AddressLen]byte(tidBytes)
		return codec.Encode(tid), nil
	case "uint8":
		n, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return nil, err
		}
		return codec.Encode[uint8](uint8(n)), nil
	case "uint16":
		n, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return nil, err
		}
		return codec.Encode[uint16](uint16(n)), nil
	case "uint32":
		n, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return nil, err
		}
		return codec.Encode[uint32](uint32(n)), nil
	case "uint64":
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return codec.Encode[uint64](n), nil
	}
	return nil, fmt.Errorf("ValueFromString: No handler for type %s", vtype)
}

//nolint:funlen,gocyclo
func ValueToString(vtype string, v []byte) (string, error) {
	switch strings.ToLower(vtype) {
	case "address":
		addr, err := codec.Decode[*cryptolib.Address](v)
		if err != nil {
			return "", err
		}
		return addr.String(), nil
	case "coinbalances":
		cbs, err := codec.Decode[*isc.CoinBalances](v)
		if err != nil {
			return "", err
		}
		return cbs.String(), nil
	case "assets":
		assets, err := codec.Decode[*isc.Assets](v)
		if err != nil {
			return "", err
		}
		return assets.String(), nil
	case "agentid":
		aid, err := codec.Decode[isc.AgentID](v)
		if err != nil {
			return "", err
		}
		return aid.String(), nil
	case "bigint":
		n, err := codec.Decode[*big.Int](v)
		if err != nil {
			return "", err
		}
		return n.String(), nil
	case "bool":
		b, err := codec.Decode[bool](v)
		if err != nil {
			return "", err
		}
		if b {
			return "true", nil
		}
		return "false", nil
	case "bytes", "hex":
		b, err := codec.Decode[[]byte](v)
		if err != nil {
			return "", err
		}
		return cryptolib.EncodeHex(b), nil
	case "chainid":
		cid, err := codec.Decode[isc.ChainID](v)
		if err != nil {
			return "", err
		}
		return cid.String(), nil
	case "dict":
		d, err := codec.Decode[dict.Dict](v)
		if err != nil {
			return "", err
		}
		s, err := d.MarshalJSON()
		if err != nil {
			return "", err
		}
		return string(s), nil
	case "hash":
		hash, err := codec.Decode[hashing.HashValue](v)
		if err != nil {
			return "", err
		}
		return hash.String(), nil
	case "hname":
		hn, err := codec.Decode[isc.Hname](v)
		if err != nil {
			return "", err
		}
		return hn.String(), nil
	case "int8":
		n, err := codec.Decode[int8](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	case "int16":
		n, err := codec.Decode[int16](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	case "int32":
		n, err := codec.Decode[int32](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	case "int64", "int":
		n, err := codec.Decode[int64](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	case "objectid":
		nid, err := codec.Decode[iotago.ObjectID](v)
		if err != nil {
			return "", err
		}
		return nid.String(), nil
	case "requestid":
		rid, err := codec.Decode[isc.RequestID](v)
		if err != nil {
			return "", err
		}
		return rid.String(), nil
	case "string":
		return fmt.Sprintf("%q", string(v)), nil
	case "tokenid":
		tid, err := codec.Decode[coin.Type](v)
		if err != nil {
			return "", err
		}
		return tid.String(), nil
	case "uint8":
		n, err := codec.Decode[uint8](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	case "uint16":
		n, err := codec.Decode[uint16](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	case "uint32":
		n, err := codec.Decode[uint32](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	case "uint64":
		n, err := codec.Decode[uint64](v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", n), nil
	}

	return "", fmt.Errorf("ValueToString: No handler for type %s", vtype)
}

func EncodeParams(params []string) (isc.CallArguments, error) {
	if len(params)%2 != 0 {
		return nil, fmt.Errorf("params format: '<type> <value> ...'")
	}

	encodedParams := make(isc.CallArguments, 0, len(params)/2)

	for i := 0; i < len(params)/2; i++ {
		vtype := params[i*2]
		v := params[i*2+1]

		val, err := ValueFromString(vtype, v)
		if err != nil {
			return nil, err
		}
		encodedParams = append(encodedParams, val)
	}

	return encodedParams, nil
}

func PrintCallResultsAsJSON(res isc.CallResults) error {
	return json.NewEncoder(os.Stdout).Encode(models.ToCallResultsJSON(res))
}

func ReadCallResultsAsJSON() (isc.CallArguments, error) {
	var args models.CallResultsJSON
	err := json.NewDecoder(os.Stdin).Decode(&args)
	if err != nil {
		return nil, err
	}
	result, err := args.ToCallResults()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func AgentIDFromArgs(args []string) (isc.AgentID, error) {
	if len(args) == 0 {
		return isc.NewAddressAgentID(wallet.Load().Address()), nil
	}
	return AgentIDFromString(args[0])
}

func AgentIDFromString(s string) (isc.AgentID, error) {
	if s == "common" {
		return accounts.CommonAccount(), nil
	}

	agentID, err := isc.AgentIDFromString(s)
	if err != nil {
		return nil, fmt.Errorf("cannot parse AgentID: %w", err)
	}
	return agentID, nil
}
