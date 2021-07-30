package rootimpl

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/_default"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract is not found by the given hname, the default contract is returned
// the bool flag indicates if contract was found or not. It means, if hname == 0 it will
// return default contract and true
func FindContract(state kv.KVStoreReader, hname iscp.Hname) (*root.ContractRecord, bool) {
	contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
	retBin := contractRegistry.MustGetAt(hname.Bytes())
	if retBin != nil {
		ret, err := root.ContractRecordFromBytes(retBin)
		if err != nil {
			panic(xerrors.Errorf("FindContract: %w", err))
		}
		return ret, true
	}
	if hname == root.Contract.Hname() {
		// it happens during bootstrap
		return root.NewContractRecord(root.Contract, &iscp.NilAgentID), true
	}
	return root.NewContractRecord(_default.Contract, &iscp.NilAgentID), false
}

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader) root.ChainInfo {
	d := kvdecoder.New(state)
	ret := root.ChainInfo{
		ChainID:             *d.MustGetChainID(root.VarChainID),
		ChainOwnerID:        *d.MustGetAgentID(root.VarChainOwnerID),
		Description:         d.MustGetString(root.VarDescription, ""),
		FeeColor:            d.MustGetColor(root.VarFeeColor, colored.IOTA),
		DefaultOwnerFee:     d.MustGetInt64(root.VarDefaultOwnerFee, 0),
		DefaultValidatorFee: d.MustGetInt64(root.VarDefaultValidatorFee, 0),
		MaxBlobSize:         d.MustGetUint32(root.VarMaxBlobSize, 0),
		MaxEventSize:        d.MustGetUint16(root.VarMaxEventSize, 0),
		MaxEventsPerReq:     d.MustGetUint16(root.VarMaxEventsPerReq, 0),
	}
	return ret
}

func MustGetChainOwnerID(state kv.KVStoreReader) *iscp.AgentID {
	d := kvdecoder.New(state)
	return d.MustGetAgentID(root.VarChainOwnerID)
}

// GetFeeInfo is an internal utility function which returns fee info for the contract
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not exposed to the sandbox
func GetFeeInfo(ctx iscp.SandboxView, hname iscp.Hname) (colored.Color, uint64, uint64) {
	state := ctx.State()
	rec, found := FindContract(state, hname)
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
	feeColor := deco.MustGetColor(root.VarFeeColor, colored.IOTA)
	defaultOwnerFee := deco.MustGetUint64(root.VarDefaultOwnerFee, 0)
	defaultValidatorFee := deco.MustGetUint64(root.VarDefaultValidatorFee, 0)
	return feeColor, defaultOwnerFee, defaultValidatorFee, nil
}

// DecodeContractRegistry encodes the whole contract registry from the map into a Go map.
func DecodeContractRegistry(contractRegistry *collections.ImmutableMap) (map[iscp.Hname]*root.ContractRecord, error) {
	ret := make(map[iscp.Hname]*root.ContractRecord)
	var err error
	contractRegistry.MustIterate(func(k []byte, v []byte) bool {
		var deploymentHash iscp.Hname
		deploymentHash, err = iscp.HnameFromBytes(k)
		if err != nil {
			return false
		}

		cr, err := root.ContractRecordFromBytes(v)
		if err != nil {
			return false
		}

		ret[deploymentHash] = cr
		return true
	})
	return ret, err
}

func CheckAuthorizationByChainOwner(state kv.KVStore, agentID *iscp.AgentID) bool {
	currentOwner, _, err := codec.DecodeAgentID(state.MustGet(root.VarChainOwnerID))
	if err != nil {
		panic(err)
	}
	return currentOwner.Equals(agentID)
}

func mustStoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo, a assert.Assert) {
	rec := root.NewContractRecord(i, &iscp.NilAgentID)
	ctx.Log().Debugf("mustStoreAndInitCoreContract: '%s', hname = %s", i.Name, i.Hname())
	mustStoreContractRecord(ctx, rec, a)
}

func mustStoreAndInitCoreContract(ctx iscp.Sandbox, i *coreutil.ContractInfo, a assert.Assert) {
	mustStoreContract(ctx, i, a)
	_, err := ctx.Call(iscp.Hn(i.Name), iscp.EntryPointInit, nil, nil)
	a.RequireNoError(err)
}

func mustStoreContractRecord(ctx iscp.Sandbox, rec *root.ContractRecord, a assert.Assert) {
	hname := rec.Hname()
	contractRegistry := collections.NewMap(ctx.State(), root.VarContractRegistry)
	a.Require(!contractRegistry.MustHasAt(hname.Bytes()), "contract '%s'/%s already exist", rec.Name, hname.String())
	contractRegistry.MustSetAt(hname.Bytes(), rec.Bytes())
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

	return collections.NewMap(ctx.State(), root.VarDeployPermissions).MustHasAt(caller.Bytes())
}
