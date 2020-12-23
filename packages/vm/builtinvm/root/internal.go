package root

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not exposed to the sandbox
func FindContract(state kv.KVStore, hname coretypes.Hname) (*ContractRecord, error) {
	contractRegistry := datatypes.NewMustMap(state, VarContractRegistry)
	retBin := contractRegistry.GetAt(hname.Bytes())
	if retBin == nil {
		if hname == Interface.Hname() {
			// if not found and it is root, it means it is chain init --> return empty root record
			rec := NewContractRecord(Interface, coretypes.AgentID{})
			return &rec, nil
		}
		return nil, fmt.Errorf("root: contract %s not found", hname)
	}

	ret, err := DecodeContractRecord(retBin)
	if err != nil {
		return nil, fmt.Errorf("root: %v", err)
	}
	return ret, nil
}

// GetChainInfo return global variables of the chain
func GetChainInfo(state kv.KVStore) (*ChainInfo, error) {
	ret := &ChainInfo{}
	var err error
	ret.ChainID, _, err = codec.DecodeChainID(state.MustGet(VarChainID))
	if err != nil {
		return nil, err
	}
	ret.ChainOwnerID, _, err = codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	if err != nil {
		return nil, err
	}
	ret.ChainColor, _, err = codec.DecodeColor(state.MustGet(VarChainColor))
	if err != nil {
		return nil, err
	}
	ret.ChainAddress, _, err = codec.DecodeAddress(state.MustGet(VarChainAddress))
	if err != nil {
		return nil, err
	}
	ret.Description, _, err = codec.DecodeString(state.MustGet(VarDescription))
	if err != nil {
		return nil, err
	}
	feeColor, ok, err := codec.DecodeColor(state.MustGet(VarFeeColor))
	if err != nil {
		return nil, err
	}
	if ok {
		ret.FeeColor = feeColor
	} else {
		ret.FeeColor = balance.ColorIOTA
	}
	defaultOwnerFee, ok, err := codec.DecodeInt64(state.MustGet(VarDefaultOwnerFee))
	if err != nil {
		return nil, err
	}
	if ok {
		ret.DefaultOwnerFee = defaultOwnerFee
	}
	defaultValidatorFee, ok, err := codec.DecodeInt64(state.MustGet(VarDefaultValidatorFee))
	if err != nil {
		return nil, err
	}
	if ok {
		ret.DefaultValidatorFee = defaultValidatorFee
	}
	return ret, nil
}

// GetFeeInfo is an internal utility function which returns fee info for the contract
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not exposed to the sandbox
func GetFeeInfo(state kv.KVStore, hname coretypes.Hname) (balance.Color, int64, int64, error) {
	rec, err := FindContract(state, hname)
	if err != nil {
		// contract not found
		return balance.Color{}, 0, 0, err
	}
	feeColor, defaultOwnerFee, defaultValidatorFee, err := GetDefaultFeeInfo(state)
	ownerFee := rec.OwnerFee
	if ownerFee == 0 {
		ownerFee = defaultOwnerFee
	}
	validatorFee := rec.ValidatorFee
	if validatorFee == 0 {
		validatorFee = defaultValidatorFee
	}
	return feeColor, ownerFee, validatorFee, nil
}

func GetDefaultFeeInfo(state kv.KVStore) (balance.Color, int64, int64, error) {
	feeColor, ok, _ := codec.DecodeColor(state.MustGet(VarFeeColor))
	if !ok {
		feeColor = balance.ColorIOTA
	}
	defaultOwnerFee, _, err := codec.DecodeInt64(state.MustGet(VarDefaultOwnerFee))
	if err != nil {
		return balance.Color{}, 0, 0, err
	}
	defaultValidatorFee, _, err := codec.DecodeInt64(state.MustGet(VarDefaultValidatorFee))
	if err != nil {
		return balance.Color{}, 0, 0, err
	}
	return feeColor, defaultOwnerFee, defaultValidatorFee, nil
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
