package root

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
func FindContract(state kv.KVStoreReader, hname coretypes.Hname) (*ContractRecord, error) {
	contractRegistry := collections.NewMapReadOnly(state, VarContractRegistry)
	retBin := contractRegistry.MustGetAt(hname.Bytes())
	if retBin == nil {
		if hname == Interface.Hname() {
			// if not found and it is root, it means it is chain init --> return empty root record
			rec := NewContractRecord(Interface, coretypes.AgentID{})
			return &rec, nil
		}
		return nil, ErrContractNotFound
	}

	ret, err := DecodeContractRecord(retBin)
	if err != nil {
		return nil, fmt.Errorf("root: %v", err)
	}
	return ret, nil
}

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader) ChainInfo {
	d := kvdecoder.New(state)
	ret := ChainInfo{
		ChainID:             d.MustGetChainID(VarChainID),
		ChainOwnerID:        d.MustGetAgentID(VarChainOwnerID),
		ChainColor:          d.MustGetColor(VarChainColor),
		ChainAddress:        d.MustGetAddress(VarChainAddress),
		Description:         d.MustGetString(VarDescription),
		FeeColor:            d.MustGetColor(VarFeeColor, balance.ColorIOTA),
		DefaultOwnerFee:     d.MustGetInt64(VarDefaultOwnerFee, 0),
		DefaultValidatorFee: d.MustGetInt64(VarDefaultValidatorFee, 0),
	}
	return ret
}

// GetFeeInfo is an internal utility function which returns fee info for the contract
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not exposed to the sandbox
func GetFeeInfo(state kv.KVStoreReader, hname coretypes.Hname) (balance.Color, int64, int64) {
	//returns nil of contract not found
	rec, err := FindContract(state, hname)
	if err != nil {
		if err != ErrContractNotFound {
			panic(err)
		} else {
			rec = nil
		}
	}
	return GetFeeInfoByContractRecord(state, rec)
}

func GetFeeInfoByContractRecord(state kv.KVStoreReader, rec *ContractRecord) (balance.Color, int64, int64) {
	var ownerFee, validatorFee int64
	if rec != nil {
		ownerFee = rec.OwnerFee
		validatorFee = rec.ValidatorFee
	}
	feeColor, defaultOwnerFee, defaultValidatorFee, err := GetDefaultFeeInfo(state)
	if err != nil {
		panic(err)
	}
	if ownerFee == 0 {
		ownerFee = defaultOwnerFee
	}
	if validatorFee == 0 {
		validatorFee = defaultValidatorFee
	}
	return feeColor, ownerFee, validatorFee
}

func GetDefaultFeeInfo(state kv.KVStoreReader) (balance.Color, int64, int64, error) {
	feeColor, ok, err := codec.DecodeColor(state.MustGet(VarFeeColor))
	if err != nil {
		panic(err)
	}
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

// DecodeContractRegistry encodes the whole contract registry from the map into a Go map.
func DecodeContractRegistry(contractRegistry *collections.ImmutableMap) (map[coretypes.Hname]*ContractRecord, error) {
	ret := make(map[coretypes.Hname]*ContractRecord)
	var err error
	contractRegistry.MustIterate(func(k []byte, v []byte) bool {
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

func CheckAuthorizationByChainOwner(state kv.KVStore, agentID coretypes.AgentID) bool {
	currentOwner, _, err := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	if err != nil {
		panic(err)
	}
	return currentOwner == agentID
}

// storeAndInitContract internal utility function
func storeAndInitContract(ctx coretypes.Sandbox, rec *ContractRecord, initParams dict.Dict) error {
	hname := coretypes.Hn(rec.Name)
	contractRegistry := collections.NewMap(ctx.State(), VarContractRegistry)
	if contractRegistry.MustHasAt(hname.Bytes()) {
		return fmt.Errorf("contract '%s'/%s already exist", rec.Name, hname.String())
	}
	contractRegistry.MustSetAt(hname.Bytes(), EncodeContractRecord(rec))
	_, err := ctx.Call(coretypes.Hn(rec.Name), coretypes.EntryPointInit, initParams, nil)
	if err != nil {
		// call to 'init' failed: delete record
		contractRegistry.MustDelAt(hname.Bytes())
		err = fmt.Errorf("contract '%s'/%s: calling 'init': %v", rec.Name, hname.String(), err)
	}
	return err
}

// isAuthorizedToDeploy checks if caller is authorized to deploy smart contract
func isAuthorizedToDeploy(ctx coretypes.Sandbox) bool {
	if ctx.Caller() == ctx.ChainOwnerID() {
		// chain owner is always authorized
		return true
	}
	if !ctx.Caller().IsAddress() {
		// smart contract from the same chain is always authorize
		return ctx.Caller().MustContractID().ChainID() == ctx.ContractID().ChainID()
	}
	return collections.NewMap(ctx.State(), VarDeployPermissions).MustHasAt(ctx.Caller().Bytes())
}
