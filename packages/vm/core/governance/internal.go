// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"golang.org/x/xerrors"
)

// GetRotationAddress tries to read the state of 'governance' and extract rotation address
// If succeeds, it means this block is fake.
// If fails, return nil
func GetRotationAddress(state kv.KVStoreReader) iotago.Address {
	ret, err := codec.DecodeAddress(state.MustGet(StateVarRotateToAddress), nil)
	if err != nil {
		return nil
	}
	return ret
}

// MustGetChainInfo return global variables of the chain
func MustGetChainInfo(state kv.KVStoreReader) ChainInfo {
	d := kvdecoder.New(state)
	ret := ChainInfo{
		ChainID:             d.MustGetChainID(VarChainID),
		ChainOwnerID:        d.MustGetAgentID(VarChainOwnerID),
		Description:         d.MustGetString(VarDescription, ""),
		FeeAssetID:          d.MustGetBytes(VarFeeAssetID, iscp.IotaAssetID),
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
// It is called from VMContext and viewcontext objects
// It is not exposed to the sandbox
func GetFeeInfo(ctx iscp.SandboxView, hname iscp.Hname) ([]byte, uint64, uint64) {
	state := ctx.State()
	rec := FindContractFees(state, hname)
	return GetFeeInfoFromContractFeesRecord(state, rec)
}

// GetFeeInfoByHname is an internal utility function which returns fee info for the contract
// It is called from VMContext and viewcontext objects
// It is not exposed to the sandbox
func GetFeeInfoByHname(state kv.KVStoreReader, hname iscp.Hname) ([]byte, uint64, uint64) {
	rec := FindContractFees(state, hname)
	return GetFeeInfoFromContractFeesRecord(state, rec)
}

// FindContractFees is an internal utility function which finds a contract in the KVStore
// It is called from within the 'governance' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract fees are not found by the given hname, nil is returned
// the bool flag indicates if a contract-fees record was found or not
func FindContractFees(state kv.KVStoreReader, hname iscp.Hname) *ContractFeesRecord {
	contractRegistry := collections.NewMapReadOnly(state, VarContractFeesRegistry)
	retBin := contractRegistry.MustGetAt(hname.Bytes())
	if retBin == nil {
		return nil
	}
	ret, err := ContractFeesRecordFromBytes(retBin)
	if err != nil {
		panic(xerrors.Errorf("FindContractFees: %w", err))
	}
	return ret
}

func GetFeeInfoFromContractFeesRecord(state kv.KVStoreReader, rec *ContractFeesRecord) ([]byte, uint64, uint64) {
	var ownerFee, validatorFee uint64
	if rec != nil {
		ownerFee = rec.OwnerFee
		validatorFee = rec.ValidatorFee
	}
	assetID, defaultOwnerFee, defaultValidatorFee, err := GetDefaultFeeInfo(state)
	if err != nil {
		panic(err)
	}
	if ownerFee == 0 {
		ownerFee = defaultOwnerFee
	}
	if validatorFee == 0 {
		validatorFee = defaultValidatorFee
	}
	return assetID, ownerFee, validatorFee
}

func GetDefaultFeeInfo(state kv.KVStoreReader) ([]byte, uint64, uint64, error) {
	deco := kvdecoder.New(state)
	feeAssetID := deco.MustGetBytes(VarFeeAssetID, iscp.IotaAssetID)
	defaultOwnerFee := deco.MustGetUint64(VarDefaultOwnerFee, 0)
	defaultValidatorFee := deco.MustGetUint64(VarDefaultValidatorFee, 0)
	return feeAssetID, defaultOwnerFee, defaultValidatorFee, nil
}

func CheckAuthorizationByChainOwner(state kv.KVStore, agentID *iscp.AgentID) bool {
	currentOwner, err := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	if err != nil {
		panic(err)
	}
	return currentOwner.Equals(agentID)
}
