// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sm_gpa_utils

import (
	"crypto/rand"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type anchorData struct {
	ref          *sui.ObjectRef
	assets       iscmove.Referent[iscmove.AssetsBag]
	l1Commitment *state.L1Commitment
	stateIndex   uint32
}

type BlockFactory struct {
	t                   require.TestingT
	store               state.Store
	chainID             isc.ChainID
	chainInitParams     isc.CallArguments
	lastBlockCommitment *state.L1Commitment
	anchorData          map[state.BlockHash]anchorData
}

type BlockFactoryCallArguments struct {
	BlockKeepAmount int32
}

func NewBlockFactory(t require.TestingT, chainInitParamsOpt ...BlockFactoryCallArguments) *BlockFactory {
	var chainInitParams isc.CallArguments
	agentId := isc.NewRandomAgentID()
	if len(chainInitParamsOpt) > 0 {
		chainInitParams = isc.NewCallArguments(agentId.Bytes(), codec.Uint16.Encode(evm.DefaultChainID), codec.Int32.Encode(int32(chainInitParamsOpt[0].BlockKeepAmount)))
	} else {
		chainInitParams = isc.NewCallArguments(agentId.Bytes())
	}
	chainID := isc.RandomChainID()
	chainStore := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	originBlock, _ := origin.InitChain(0, chainStore, chainInitParams, 0, isc.BaseTokenCoinInfo)
	originCommitment := originBlock.L1Commitment()
	originAnchorData := anchorData{
		ref: &sui.ObjectRef{
			ObjectID: sui.ObjectIDFromArray(chainID),
			Version:  0,
			Digest:   nil, // TODO
		},
		assets: iscmove.Referent[iscmove.AssetsBag]{
			// ID: nil, // TODO
			Value: &iscmove.AssetsBag{
				// ID:   nil, // TODO
				Size: 0,
			},
		},
		l1Commitment: originCommitment,
		stateIndex:   0,
	}
	return &BlockFactory{
		t:                   t,
		store:               chainStore,
		chainID:             chainID,
		chainInitParams:     chainInitParams,
		lastBlockCommitment: originCommitment,
		anchorData:          map[state.BlockHash]anchorData{originCommitment.BlockHash(): originAnchorData},
	}

	/*
	   aliasOutput0ID := iotago.OutputIDFromTransactionIDAndIndex(getRandomTxID(t), 0)
	   chainID := isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(aliasOutput0ID))
	   stateAddress := cryptolib.NewKeyPair().GetPublicKey().AsAddress()
	   _ = stateAddress
	   originCommitment := origin.L1Commitment(0, chainInitParams, 0)

	   	aliasOutput0 := &iotago.AliasOutput{
	   		Amount:        tpkg.TestTokenSupply,
	   		AliasID:       chainID.AsAliasID(), // NOTE: not very correct: origin output's AliasID should be empty; left here to make mocking transitions easier
	   		StateMetadata: testutil.DummyStateMetadata(originCommitment).Bytes(),
	   		Conditions:    iotago.UnlockConditions{
	   			// &iotago.StateControllerAddressUnlockCondition{Address: stateAddress.AsIotagoAddress()},
	   			// &iotago.GovernorAddressUnlockCondition{Address: stateAddress.AsIotagoAddress()},
	   		},
	   		Features: iotago.Features{
	   			&iotago.SenderFeature{
	   				// Address: stateAddress.AsIotagoAddress(),
	   			},
	   		},
	   	}

	   aliasOutputs := make(map[state.BlockHash]*isc.StateAnchor)
	   originOutput := isc.NewAliasOutputWithID(aliasOutput0, aliasOutput0ID)
	   aliasOutputs[originCommitment.BlockHash()] = originOutput
	   chainStore := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	   origin.InitChain(0, chainStore, chainInitParams, 0)

	   	return &BlockFactory{
	   		t:                   t,
	   		store:               chainStore,
	   		chainID:             chainID,
	   		chainInitParams:     chainInitParams,
	   		lastBlockCommitment: originCommitment,
	   		anchors: 	         aliasOutputs,
	   	}
	*/
}

func (bfT *BlockFactory) GetChainID() isc.ChainID {
	return bfT.chainID
}

func (bfT *BlockFactory) GetChainInitParameters() isc.CallArguments {
	return bfT.chainInitParams
}

func (bfT *BlockFactory) GetOriginAnchor() *isc.StateAnchor {
	return bfT.GetAnchor(origin.L1Commitment(0, bfT.chainInitParams, 0, isc.BaseTokenCoinInfo))
}

func (bfT *BlockFactory) GetOriginBlock() state.Block {
	block, err := bfT.store.BlockByTrieRoot(origin.L1Commitment(0, bfT.chainInitParams, 0, isc.BaseTokenCoinInfo).TrieRoot())
	require.NoError(bfT.t, err)
	return block
}

func (bfT *BlockFactory) GetBlocks(
	count,
	branchingFactor int,
) []state.Block {
	blocks := bfT.GetBlocksFrom(count, branchingFactor, bfT.lastBlockCommitment)
	require.Equal(bfT.t, count, len(blocks))
	bfT.lastBlockCommitment = blocks[count-1].L1Commitment()
	return blocks
}

func (bfT *BlockFactory) GetBlocksFrom(
	count,
	branchingFactor int,
	commitment *state.L1Commitment,
	incrementFactorOpt ...uint64,
) []state.Block {
	var incrementFactor uint64
	if len(incrementFactorOpt) > 0 {
		incrementFactor = incrementFactorOpt[0]
	} else {
		incrementFactor = 1
	}
	result := make([]state.Block, count+1)
	var err error
	result[0], err = bfT.store.BlockByTrieRoot(commitment.TrieRoot())
	require.NoError(bfT.t, err)
	for i := 1; i < len(result); i++ {
		baseIndex := (i + branchingFactor - 2) / branchingFactor
		increment := uint64(1+i%branchingFactor) * incrementFactor
		result[i] = bfT.GetNextBlock(result[baseIndex].L1Commitment(), increment)
	}
	return result[1:]
}

func (bfT *BlockFactory) GetNextBlock(
	commitment *state.L1Commitment,
	incrementOpt ...uint64,
) state.Block {
	stateDraft, err := bfT.store.NewStateDraft(time.Now(), commitment)
	require.NoError(bfT.t, err)
	counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")
	counterBin := stateDraft.Get(counterKey)
	counter, err := codec.Uint64.Decode(counterBin, 0)
	require.NoError(bfT.t, err)
	var increment uint64
	if len(incrementOpt) > 0 {
		increment = incrementOpt[0]
	} else {
		increment = 1
	}
	counterBin = codec.Uint64.Encode(counter + increment)
	stateDraft.Mutations().Set(counterKey, counterBin)
	block := bfT.store.Commit(stateDraft)
	// require.EqualValues(t, stateDraft.BlockIndex(), block.BlockIndex())
	newCommitment := block.L1Commitment()

	consumedAnchor := bfT.GetAnchor(commitment)

	newAnchorData := anchorData{
		ref: &sui.ObjectRef{
			ObjectID: consumedAnchor.GetObjectID(),
			Version:  consumedAnchor.GetObjectRef().Version + 1,
			Digest:   nil, // TODO
		},
		assets:       consumedAnchor.GetAssets(),
		l1Commitment: newCommitment,
		stateIndex:   consumedAnchor.GetStateIndex() + 1,
	}
	bfT.anchorData[newCommitment.BlockHash()] = newAnchorData

	return block
}

func (bfT *BlockFactory) GetStore() state.Store {
	return NewReadOnlyStore(bfT.store)
}

func (bfT *BlockFactory) GetStateDraft(block state.Block) state.StateDraft {
	result, err := bfT.store.NewEmptyStateDraft(block.PreviousL1Commitment())
	require.NoError(bfT.t, err)
	block.Mutations().ApplyTo(result)
	return result
}

func (bfT *BlockFactory) GetAnchor(commitment *state.L1Commitment) *isc.StateAnchor {
	anchorData, ok := bfT.anchorData[commitment.BlockHash()]
	require.True(bfT.t, ok)

	metadata := transaction.StateMetadata{
		SchemaVersion: 0,
		L1Commitment:  commitment,
		GasFeePolicy:  gas.DefaultFeePolicy(),
		InitParams:    bfT.chainInitParams,
		PublicURL:     "",
	}

	return &isc.StateAnchor{
		Ref: &iscmove.RefWithObject[iscmove.Anchor]{
			ObjectRef: *anchorData.ref,
			Object: &iscmove.Anchor{
				ID:            *anchorData.ref.ObjectID,
				Assets:        anchorData.assets,
				StateMetadata: metadata.Bytes(),
				StateIndex:    anchorData.stateIndex,
			},
		},
		Owner:      nil,                            //FIXME
		ISCPackage: *sui.MustAddressFromHex("0x0"), //FIXME,
	}
}

func getRandomTxID(t require.TestingT) iotago.TransactionID {
	var result iotago.TransactionID
	_, err := rand.Read(result[:])
	require.NoError(t, err)
	return result
}
