package root

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv/codec"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not exposed to the sandbox
func FindContract(state kv.KVStore, hname coretypes.Hname) (*ContractRecord, error) {
	if hname == Interface.Hname() {
		return &RootContractRecord, nil
	}
	contractRegistry := datatypes.NewMustMap(state, VarContractRegistry)
	retBin := contractRegistry.GetAt(hname.Bytes())
	if retBin == nil {
		return nil, fmt.Errorf("root: contract %s not found", hname)
	}
	ret, err := DecodeContractRecord(retBin)
	if err != nil {
		return nil, fmt.Errorf("root: %v", err)
	}
	return ret, nil
}

// GetFeeInfo is an internal utility function which return fee info for the contract
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not exposed to the sandbox
func GetFeeInfo(state kv.KVStore, hname coretypes.Hname) (*balance.Color, int64, error) {
	rec, err := FindContract(state, hname)
	if err != nil {
		// contract not found
		return nil, 0, err
	}
	feeColor, ok, _ := codec.DecodeColor(state.MustGet(VarFeeColor))
	if !ok {
		feeColor = &balance.ColorIOTA
	}
	fee := rec.Fee
	if fee == 0 {
		// look up for default fee on chain
		defaultFee, ok, _ := codec.DecodeInt64(state.MustGet(VarDefaultFee))
		if ok {
			fee = defaultFee
		}
	}
	return feeColor, fee, nil
}

// DecodeContractRegistry encodes the whole contract registry in the MustMap in the kvstor to the
// Go map.
func DecodeContractRegistry(contractRegistry *datatypes.MustMap) (map[coretypes.Hname]*ContractRecord, error) {
	ret := make(map[coretypes.Hname]*ContractRecord)
	var err error
	contractRegistry.Iterate(func(k []byte, v []byte) bool {
		var deploymentHash coretypes.Hname
		deploymentHash, err = coretypes.NewHnameFromBytes(k)
		if err != nil {
			return false
		}

		var cr *ContractRecord
		cr, err = DecodeContractRecord(v)
		if err != nil {
			return false
		}

		ret[deploymentHash] = cr
		return true
	})
	return ret, err
}

func CheckAuthorization(state kv.KVStore, agentID coretypes.AgentID) bool {
	currentOwner, _, _ := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	return currentOwner == agentID
}
