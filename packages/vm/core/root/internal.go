package root

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp/color"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/_default"
)

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract is not found by the given hname, the default contract is returned
func FindContract(state kv.KVStoreReader, hname iscp.Hname) (*ContractRecord, error) {
	contractRegistry := collections.NewMapReadOnly(state, VarContractRegistry)
	retBin := contractRegistry.MustGetAt(hname.Bytes())
	var ret *ContractRecord
	var err error
	if retBin != nil {
		if ret, err = DecodeContractRecord(retBin); err != nil {
			return nil, fmt.Errorf("root: %v", err)
		}
	} else {
		// not founc in registry
		if hname == Contract.Hname() {
			// if not found and it is root, it means it is chain init --> return empty root record
			ret = NewContractRecord(Contract, &iscp.AgentID{})
		} else {
			// return default contract
			ret = NewContractRecord(_default.Contract, &iscp.AgentID{})
		}
	}
	return ret, nil
}

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader) ChainInfo {
	d := kvdecoder.New(state)
	ret := ChainInfo{
		ChainID:             *d.MustGetChainID(VarChainID),
		ChainOwnerID:        *d.MustGetAgentID(VarChainOwnerID),
		Description:         d.MustGetString(VarDescription, ""),
		FeeColor:            d.MustGetColorLedgerstate(VarFeeColor, ledgerstate.ColorIOTA),
		DefaultOwnerFee:     d.MustGetInt64(VarDefaultOwnerFee, 0),
		DefaultValidatorFee: d.MustGetInt64(VarDefaultValidatorFee, 0),
	}
	return ret
}

func MustGetChainOwnerID(state kv.KVStoreReader) *iscp.AgentID {
	d := kvdecoder.New(state)
	return d.MustGetAgentID(VarChainOwnerID)
}

// GetFeeInfo is an internal utility function which returns fee info for the contract
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not exposed to the sandbox
func GetFeeInfo(state kv.KVStoreReader, hname iscp.Hname) (color.Color, uint64, uint64) {
	// returns nil of contract not found
	rec, err := FindContract(state, hname)
	if err != nil {
		if !errors.Is(err, ErrContractNotFound) {
			panic(err)
		} else {
			rec = nil
		}
	}
	return GetFeeInfoByContractRecord(state, rec)
}

func GetFeeInfoByContractRecord(state kv.KVStoreReader, rec *ContractRecord) (color.Color, uint64, uint64) {
	var ownerFee, validatorFee uint64
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

func GetDefaultFeeInfo(state kv.KVStoreReader) (color.Color, uint64, uint64, error) {
	deco := kvdecoder.New(state)
	feeColor := deco.MustGetColor(VarFeeColor, color.IOTA)
	defaultOwnerFee := deco.MustGetUint64(VarDefaultOwnerFee, 0)
	defaultValidatorFee := deco.MustGetUint64(VarDefaultValidatorFee, 0)
	return feeColor, defaultOwnerFee, defaultValidatorFee, nil
}

// DecodeContractRegistry encodes the whole contract registry from the map into a Go map.
func DecodeContractRegistry(contractRegistry *collections.ImmutableMap) (map[iscp.Hname]*ContractRecord, error) {
	ret := make(map[iscp.Hname]*ContractRecord)
	var err error
	contractRegistry.MustIterate(func(k []byte, v []byte) bool {
		var deploymentHash iscp.Hname
		deploymentHash, err = iscp.HnameFromBytes(k)
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

func CheckAuthorizationByChainOwner(state kv.KVStore, agentID *iscp.AgentID) bool {
	currentOwner, _, err := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	if err != nil {
		panic(err)
	}
	return currentOwner.Equals(agentID)
}

func mustStoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo, a assert.Assert) {
	rec := NewContractRecord(i, &iscp.AgentID{})
	ctx.Log().Debugf("mustStoreAndInitCoreContract: '%s', hname = %s", i.Name, i.Hname())
	mustStoreContractRecord(ctx, rec, a)
}

func mustStoreAndInitCoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo, a assert.Assert) {
	mustStoreContract(ctx, i, a)
	_, err := ctx.Call(iscp.Hn(i.Name), iscp.EntryPointInit, nil, nil)
	a.RequireNoError(err)
}

func mustStoreContractRecord(ctx iscp.Sandbox, rec *ContractRecord, a assert.Assert) {
	hname := rec.Hname()
	contractRegistry := collections.NewMap(ctx.State(), VarContractRegistry)
	a.Require(!contractRegistry.MustHasAt(hname.Bytes()), "contract '%s'/%s already exist", rec.Name, hname.String())
	contractRegistry.MustSetAt(hname.Bytes(), EncodeContractRecord(rec))
}

// isAuthorizedToDeploy checks if caller is authorized to deploy smart contract
func isAuthorizedToDeploy(ctx iscp.Sandbox) bool {
	caller := ctx.Caller()
	if caller.Equals(ctx.ChainOwnerID()) {
		// chain owner is always authorized
		return true
	}
	if caller.Address().Equals(ctx.ChainID().AsAddress()) {
		// smart contract from the same chain is always authorize
		return true
	}

	return collections.NewMap(ctx.State(), VarDeployPermissions).MustHasAt(caller.Bytes())
}
