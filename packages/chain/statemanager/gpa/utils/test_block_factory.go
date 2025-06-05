// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"time"

	"fortio.org/safecast"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type BlockFactory struct {
	t                   require.TestingT
	store               state.Store
	chainID             isc.ChainID
	chainInitParams     isc.CallArguments
	lastBlockCommitment *state.L1Commitment
	anchorData          map[state.BlockHash]iscmove.AnchorWithRef
}

type BlockFactoryCallArguments struct {
	BlockKeepAmount int32
}

func NewBlockFactory(t require.TestingT, chainInitParamsOpt ...BlockFactoryCallArguments) *BlockFactory {
	var chainInitParams isc.CallArguments
	agentID := isctest.NewRandomAgentID()
	if len(chainInitParamsOpt) > 0 {
		initParams := origin.DefaultInitParams(agentID)
		initParams.BlockKeepAmount = chainInitParamsOpt[0].BlockKeepAmount
		chainInitParams = initParams.Encode()
	} else {
		chainInitParams = origin.DefaultInitParams(agentID).Encode()
	}

	chainID := isctest.RandomChainID()
	chainIDObjID := chainID.AsObjectID()
	chainStore := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	originBlock, _ := origin.InitChain(allmigrations.LatestSchemaVersion, chainStore, chainInitParams, iotago.ObjectID{}, 0, parameterstest.L1Mock)
	originCommitment := originBlock.L1Commitment()
	originStateMetadata := transaction.NewStateMetadata(
		allmigrations.LatestSchemaVersion,
		originBlock.L1Commitment(),
		&iotago.ObjectID{},
		gas.DefaultFeePolicy(),
		chainInitParams,
		0,
		"",
	)
	originAnchorData := iscmove.RefWithObject[iscmove.Anchor]{
		ObjectRef: iotago.ObjectRef{
			ObjectID: &chainIDObjID,
			Version:  0,
			Digest:   nil,
		},
		Object: &iscmove.Anchor{
			ID: chainIDObjID,
			Assets: iscmove.Referent[iscmove.AssetsBag]{
				ID:    *iotatest.RandomAddress(),
				Value: lo.ToPtr(iscmovetest.RandomAssetsBag()),
			},
			StateMetadata: originStateMetadata.Bytes(),
			StateIndex:    0,
		},
	}
	return &BlockFactory{
		t:                   t,
		store:               chainStore,
		chainID:             chainID,
		chainInitParams:     chainInitParams,
		lastBlockCommitment: originCommitment,
		anchorData:          map[state.BlockHash]iscmove.RefWithObject[iscmove.Anchor]{originCommitment.BlockHash(): originAnchorData},
	}

	/*
	   aliasOutput0ID := iotago.OutputIDFromTransactionIDAndIndex(getRandomTxID(t), 0)
	   chainID := isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(aliasOutput0ID))
	   stateAddress := cryptolib.NewKeyPair().GetPublicKey().AsAddress()
	   _ = stateAddress
	   originCommitment := origin.L1Commitment(allmigrations.LatestSchemaVersion, chainInitParams, 0)

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
	return bfT.GetAnchor(origin.L1Commitment(allmigrations.LatestSchemaVersion, bfT.chainInitParams, iotago.ObjectID{}, 0, parameterstest.L1Mock))
}

func (bfT *BlockFactory) GetOriginBlock() state.Block {
	block, err := bfT.store.BlockByTrieRoot(origin.L1Commitment(allmigrations.LatestSchemaVersion, bfT.chainInitParams, iotago.ObjectID{}, 0, parameterstest.L1Mock).TrieRoot())
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
		val := 1 + i%branchingFactor
		increment64, err := safecast.Convert[uint64](val)
		require.NoError(bfT.t, err)
		increment := increment64 * incrementFactor
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
	counter, err := codec.Decode[uint64](counterBin, 0)
	require.NoError(bfT.t, err)
	var increment uint64
	if len(incrementOpt) > 0 {
		increment = incrementOpt[0]
	} else {
		increment = 1
	}
	counterBin = codec.Encode(counter + increment)
	stateDraft.Mutations().Set(counterKey, counterBin)
	block := bfT.store.Commit(stateDraft)
	newCommitment := block.L1Commitment()

	consumedAnchor := bfT.GetAnchor(commitment)
	consumedMetadata, err := transaction.StateMetadataFromBytes(consumedAnchor.Anchor().Object.StateMetadata)
	require.NoError(bfT.t, err)
	consumedMetadata.L1Commitment, err = state.NewL1CommitmentFromBytes(newCommitment.Bytes())
	require.NoError(bfT.t, err)

	newAnchor := isc.NewStateAnchor(&iscmove.AnchorWithRef{
		Owner:     consumedAnchor.Anchor().Owner,
		ObjectRef: consumedAnchor.Anchor().ObjectRef,
		Object: &iscmove.Anchor{
			ID:            consumedAnchor.Anchor().Object.ID,
			Assets:        consumedAnchor.Anchor().Object.Assets,
			StateMetadata: consumedMetadata.Bytes(),
			StateIndex:    consumedAnchor.Anchor().Object.StateIndex + 1,
		},
	}, consumedAnchor.ISCPackage())

	bfT.anchorData[newCommitment.BlockHash()] = *newAnchor.Anchor()

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
	anchor, ok := bfT.anchorData[commitment.BlockHash()]
	require.True(bfT.t, ok)

	stateAnchor := isc.NewStateAnchor(&anchor, *iotago.MustAddressFromHex("0x0"))

	return &stateAnchor
}
