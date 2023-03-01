package vmcontext

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

type StateMetadata struct {
	L1Commitment   *state.L1Commitment
	GasFeePolicy   *gas.GasFeePolicy
	SchemaVersion  uint32
	CustomMetadata string
}

func (s *StateMetadata) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBytes(s.L1Commitment.Bytes())
	mu.WriteUint32(s.SchemaVersion)
	mu.WriteBytes(s.GasFeePolicy.Bytes())
	mu.WriteUint8(uint8(len(s.CustomMetadata)))
	mu.WriteBytes([]byte(s.CustomMetadata))
	return mu.Bytes()
}

func StateMetadataFromBytes(data []byte) (*StateMetadata, error) {
	ret := &StateMetadata{}
	mu := marshalutil.New(data)
	l1CommitmentBytes, err := mu.ReadBytes(state.L1CommitmentSize)
	if err != nil {
		return nil, err
	}
	ret.L1Commitment, err = state.L1CommitmentFromBytes(l1CommitmentBytes)
	if err != nil {
		return nil, err
	}
	ret.SchemaVersion, err = mu.ReadUint32()
	if err != nil {
		return nil, err
	}
	ret.GasFeePolicy, err = gas.FeePolicyFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	customMetadataLength, err := mu.ReadUint8()
	if err != nil {
		return nil, err
	}
	customMetadataBytes, err := mu.ReadBytes(int(customMetadataLength))
	if err != nil {
		return nil, err
	}
	ret.CustomMetadata = string(customMetadataBytes)
	return ret, nil
}

func (vmctx *VMContext) StateMetadata(stateCommitment *state.L1Commitment) []byte {
	stateMetadata := StateMetadata{
		L1Commitment: stateCommitment,
	}
	if vmctx.currentStateUpdate == nil {
		// create a temporary empty state update, so that vmctx.callCore works
		vmctx.currentStateUpdate = NewStateUpdate()
		defer func() { vmctx.currentStateUpdate = nil }()
	}

	vmctx.callCore(root.Contract, func(s kv.KVStore) {
		stateMetadata.SchemaVersion = root.GetSchemaVersion(s)
	})

	vmctx.callCore(governance.Contract, func(s kv.KVStore) {
		stateMetadata.CustomMetadata = governanceimpl.GetCustomMetadata(s)
		stateMetadata.GasFeePolicy = governance.MustGetGasFeePolicy(s)
	})

	return stateMetadata.Bytes()
}

func (vmctx *VMContext) BuildTransactionEssence(stateCommitment *state.L1Commitment) (*iotago.TransactionEssence, []byte) {
	stateMetadata := vmctx.StateMetadata(stateCommitment)
	return vmctx.txbuilder.BuildTransactionEssence(stateMetadata)
}

// CalcTransactionSubEssenceHash builds transaction essence from tx builder
// data assuming all zeroes in the L1 commitment. Returns hash of it.
// It is needed for fraud proofs
func (vmctx *VMContext) CalcTransactionSubEssenceHash() blocklog.TransactionEssenceHash {
	stateMetadata := vmctx.StateMetadata(state.L1CommitmentNil)
	essence, _ := vmctx.txbuilder.BuildTransactionEssence(stateMetadata)
	return blocklog.CalcTransactionEssenceHash(essence)
}

func (vmctx *VMContext) createTxBuilderSnapshot() *vmtxbuilder.AnchorTransactionBuilder {
	return vmctx.txbuilder.Clone()
}

func (vmctx *VMContext) restoreTxBuilderSnapshot(snapshot *vmtxbuilder.AnchorTransactionBuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *VMContext) loadNativeTokenOutput(nativeTokenID iotago.NativeTokenID) (*iotago.BasicOutput, iotago.OutputID) {
	var retOut *iotago.BasicOutput
	var blockIndex uint32
	var outputIndex uint16
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		retOut, blockIndex, outputIndex = accounts.GetNativeTokenOutput(s, nativeTokenID, vmctx.ChainID())
	})
	if retOut == nil {
		return nil, iotago.OutputID{}
	}

	outputID, err := vmctx.getOutputID(blockIndex, outputIndex)
	if err != nil {
		panic(fmt.Errorf("internal: can't find UTXO input for block index %d, output index %d, error: %w", blockIndex, outputIndex, err))
	}

	return retOut, outputID
}

func (vmctx *VMContext) loadFoundry(serNum uint32) (*iotago.FoundryOutput, iotago.OutputID) {
	var foundryOutput *iotago.FoundryOutput
	var blockIndex uint32
	var outputIndex uint16
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		foundryOutput, blockIndex, outputIndex = accounts.GetFoundryOutput(s, serNum, vmctx.ChainID())
	})
	if foundryOutput == nil {
		return nil, iotago.OutputID{}
	}

	outputID, err := vmctx.getOutputID(blockIndex, outputIndex)
	if err != nil {
		panic(fmt.Errorf("internal: can't find UTXO input for block index %d, output index %d, error: %w", blockIndex, outputIndex, err))
	}

	return foundryOutput, outputID
}

func (vmctx *VMContext) getOutputID(blockIndex uint32, outputIndex uint16) (iotago.OutputID, error) {
	var outputID iotago.OutputID
	var err error
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		outputID, err = blocklog.GetOutputID(s, blockIndex, outputIndex)
	})
	if err != nil {
		return iotago.OutputID{}, err
	}

	return outputID, nil
}

func (vmctx *VMContext) loadNFT(id iotago.NFTID) (*iotago.NFTOutput, iotago.OutputID) {
	var nftOutput *iotago.NFTOutput
	var blockIndex uint32
	var outputIndex uint16
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		nftOutput, blockIndex, outputIndex = accounts.GetNFTOutput(s, id, vmctx.ChainID())
	})
	if nftOutput == nil {
		return nil, iotago.OutputID{}
	}

	outputID, err := vmctx.getOutputID(blockIndex, outputIndex)
	if err != nil {
		panic(fmt.Errorf("internal: can't find UTXO input for block index %d, output index %d, error: %w", blockIndex, outputIndex, err))
	}

	return nftOutput, outputID
}
