// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"crypto/rand"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
)

type BlockFactory struct {
	t                   require.TestingT
	store               state.Store
	chainID             *isc.ChainID
	originOutput        *isc.AliasOutputWithID
	lastBlockCommitment *state.L1Commitment
	lastAliasOutput     *isc.AliasOutputWithID
}

func NewBlockFactory(t require.TestingT) *BlockFactory {
	aliasOutput0ID := iotago.OutputIDFromTransactionIDAndIndex(getRandomTxID(t), 0)
	chainID := isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(aliasOutput0ID))
	stateAddress := cryptolib.NewKeyPair().GetPublicKey().AsEd25519Address()
	aliasOutput0 := &iotago.AliasOutput{
		Amount:        tpkg.TestTokenSupply,
		AliasID:       *chainID.AsAliasID(), // NOTE: not very correct: origin output's AliasID should be empty; left here to make mocking transitions easier
		StateMetadata: state.OriginL1Commitment().Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateAddress},
			&iotago.GovernorAddressUnlockCondition{Address: stateAddress},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: stateAddress,
			},
		},
	}
	originOutput := isc.NewAliasOutputWithID(aliasOutput0, aliasOutput0ID)
	return &BlockFactory{
		t:                   t,
		store:               state.InitChainStore(mapdb.NewMapDB()),
		chainID:             &chainID,
		originOutput:        originOutput,
		lastBlockCommitment: state.OriginL1Commitment(),
		lastAliasOutput:     originOutput,
	}
}

func (bfT *BlockFactory) GetChainID() *isc.ChainID {
	return bfT.chainID
}

func (bfT *BlockFactory) GetOriginOutput() *isc.AliasOutputWithID {
	return bfT.originOutput
}

func (bfT *BlockFactory) GetBlocks(
	count,
	branchingFactor int,
) ([]state.Block, []*isc.AliasOutputWithID) {
	blocks, aliasOutpus := bfT.GetBlocksFrom(count, branchingFactor, bfT.lastBlockCommitment, bfT.lastAliasOutput)
	require.Equal(bfT.t, count, len(blocks))
	require.Equal(bfT.t, count, len(aliasOutpus))
	bfT.lastBlockCommitment = blocks[count-1].L1Commitment()
	bfT.lastAliasOutput = aliasOutpus[count-1]
	return blocks, aliasOutpus
}

func (bfT *BlockFactory) GetBlocksFrom(
	count,
	branchingFactor int,
	commitment *state.L1Commitment,
	aliasOutput *isc.AliasOutputWithID,
	incrementFactorOpt ...uint64,
) ([]state.Block, []*isc.AliasOutputWithID) {
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
	aliasOutputs := make([]*isc.AliasOutputWithID, len(result))
	aliasOutputs[0] = aliasOutput
	for i := 1; i < len(result); i++ {
		baseIndex := (i + branchingFactor - 2) / branchingFactor
		increment := uint64(1+i%branchingFactor) * incrementFactor
		result[i], aliasOutputs[i] = bfT.GetNextBlock(result[baseIndex].L1Commitment(), aliasOutputs[baseIndex], increment)
	}
	return result[1:], aliasOutputs[1:]
}

func (bfT *BlockFactory) GetNextBlock(
	commitment *state.L1Commitment,
	consumedAliasOutputWithID *isc.AliasOutputWithID,
	incrementOpt ...uint64,
) (state.Block, *isc.AliasOutputWithID) {
	stateDraft, err := bfT.store.NewStateDraft(time.Now(), commitment)
	require.NoError(bfT.t, err)
	counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")
	counterBin, err := stateDraft.Get(counterKey)
	require.NoError(bfT.t, err)
	counter, err := codec.DecodeUint64(counterBin, 0)
	require.NoError(bfT.t, err)
	var increment uint64
	if len(incrementOpt) > 0 {
		increment = incrementOpt[0]
	} else {
		increment = 1
	}
	counterBin = codec.EncodeUint64(counter + increment)
	stateDraft.Mutations().Set(counterKey, counterBin)
	block := bfT.store.Commit(stateDraft)
	//require.EqualValues(t, stateDraft.BlockIndex(), block.BlockIndex())

	consumedAliasOutput := consumedAliasOutputWithID.GetAliasOutput()
	aliasOutput := &iotago.AliasOutput{
		Amount:         consumedAliasOutput.Amount,
		NativeTokens:   consumedAliasOutput.NativeTokens,
		AliasID:        consumedAliasOutput.AliasID,
		StateIndex:     consumedAliasOutput.StateIndex + 1,
		StateMetadata:  block.L1Commitment().Bytes(),
		FoundryCounter: consumedAliasOutput.FoundryCounter,
		Conditions:     consumedAliasOutput.Conditions,
		Features:       consumedAliasOutput.Features,
	}
	aliasOutputID := iotago.OutputIDFromTransactionIDAndIndex(getRandomTxID(bfT.t), 0)
	aliasOutputWithID := isc.NewAliasOutputWithID(aliasOutput, aliasOutputID)

	return block, aliasOutputWithID
}

func (bfT *BlockFactory) GetStateDraft(block state.Block) state.StateDraft {
	result, err := bfT.store.NewEmptyStateDraft(block.PreviousL1Commitment())
	require.NoError(bfT.t, err)
	block.Mutations().ApplyTo(result)
	return result
}

func getRandomTxID(t require.TestingT) iotago.TransactionID {
	var result iotago.TransactionID
	_, err := rand.Read(result[:])
	require.NoError(t, err)
	return result
}
