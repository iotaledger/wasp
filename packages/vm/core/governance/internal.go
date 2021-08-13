package governance

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// GetRotationAddress tries to read the state of 'governance' and extract rotation address
// If succeeds, it means this block is fake.
// If fails, return nil
func GetRotationAddress(state kv.KVStoreReader) ledgerstate.Address {
	ret, ok, err := codec.DecodeAddress(state.MustGet(StateVarRotateToAddress))
	if !ok || err != nil {
		return nil
	}
	return ret
}

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader) ChainInfo {
	d := kvdecoder.New(state)
	ret := ChainInfo{
		ChainID:             *d.MustGetChainID(VarChainID),
		ChainOwnerID:        *d.MustGetAgentID(VarChainOwnerID),
		Description:         d.MustGetString(VarDescription, ""),
		FeeColor:            d.MustGetColor(VarFeeColor, colored.IOTA),
		DefaultOwnerFee:     d.MustGetInt64(VarDefaultOwnerFee, 0),
		DefaultValidatorFee: d.MustGetInt64(VarDefaultValidatorFee, 0),
		MaxBlobSize:         d.MustGetUint32(VarMaxBlobSize, 0),
		MaxEventSize:        d.MustGetUint16(VarMaxEventSize, 0),
		MaxEventsPerReq:     d.MustGetUint16(VarMaxEventsPerReq, 0),
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
func GetFeeInfo(ctx iscp.SandboxView, hname iscp.Hname) (colored.Color, uint64, uint64) {
	state := ctx.State()
	rec, found := root.FindContract(state, hname)
	assert.NewAssert(ctx.Log()).Require(found, "contract not found")
	return GetFeeInfoByContractRecord(state, rec)
}

func GetFeeInfoByContractRecord(state kv.KVStoreReader, rec *root.ContractRecord) (colored.Color, uint64, uint64) {
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

func GetDefaultFeeInfo(state kv.KVStoreReader) (colored.Color, uint64, uint64, error) {
	deco := kvdecoder.New(state)
	feeColor := deco.MustGetColor(VarFeeColor, colored.IOTA)
	defaultOwnerFee := deco.MustGetUint64(VarDefaultOwnerFee, 0)
	defaultValidatorFee := deco.MustGetUint64(VarDefaultValidatorFee, 0)
	return feeColor, defaultOwnerFee, defaultValidatorFee, nil
}

func CheckAuthorizationByChainOwner(state kv.KVStore, agentID *iscp.AgentID) bool {
	currentOwner, _, err := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	if err != nil {
		panic(err)
	}
	return currentOwner.Equals(agentID)
}
