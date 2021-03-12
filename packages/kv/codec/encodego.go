package codec

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

func Encode(v interface{}) []byte {
	switch vt := v.(type) {
	case int:
		return EncodeInt64(int64(vt))
	case byte:
		return EncodeInt64(int64(vt))
	case int16:
		return EncodeInt64(int64(vt))
	case int32:
		return EncodeInt64(int64(vt))
	case int64:
		return EncodeInt64(int64(vt))
	case uint16:
		return EncodeInt64(int64(vt))
	case uint32:
		return EncodeInt64(int64(vt))
	case uint64:
		return EncodeInt64(int64(vt))
	case string:
		return EncodeString(vt)
	case []byte:
		return vt
	case *hashing.HashValue:
		return EncodeHashValue(*vt)
	case hashing.HashValue:
		return EncodeHashValue(vt)
	case *address.Address:
		return EncodeAddress(*vt)
	case address.Address:
		return EncodeAddress(vt)
	case *balance.Color:
		return EncodeColor(*vt)
	case balance.Color:
		return EncodeColor(vt)
	case *coretypes.ChainID:
		return EncodeChainID(*vt)
	case coretypes.ChainID:
		return EncodeChainID(vt)
	case *coretypes.ContractID:
		return EncodeContractID(*vt)
	case coretypes.ContractID:
		return EncodeContractID(vt)
	case *coretypes.AgentID:
		return EncodeAgentID(*vt)
	case coretypes.AgentID:
		return EncodeAgentID(vt)
	case coretypes.Hname:
		return vt.Bytes()

	default:
		panic(fmt.Sprintf("Can't encode value %v", v))
	}
}
