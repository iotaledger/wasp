// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"math/rand"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
)

func GetOriginState(t require.TestingT) (*isc.ChainID, *isc.AliasOutputWithID, state.VirtualStateAccess) {
	store := mapdb.NewMapDB()
	aliasOutput0ID := iotago.OutputIDFromTransactionIDAndIndex(getRandomTxID(), 0).UTXOInput()
	chainID := isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(aliasOutput0ID.ID()))
	originVS, err := state.CreateOriginState(store, &chainID)
	require.NoError(t, err)
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
	return &chainID, isc.NewAliasOutputWithID(aliasOutput0, aliasOutput0ID), originVS
}

func GetBlocks(
	t require.TestingT,
	count,
	branchingFactor int,
) (*isc.ChainID, []state.Block, []*isc.AliasOutputWithID, []state.VirtualStateAccess) {
	chainID, aliasOutput0, originVS := GetOriginState(t)
	blocks, aliasOutputs, virtualStates := GetBlocksFrom(t, count, branchingFactor, originVS, aliasOutput0)
	return chainID, blocks, aliasOutputs, virtualStates
}

func GetBlocksFrom(
	t require.TestingT,
	count,
	branchingFactor int,
	vs state.VirtualStateAccess,
	aliasOutput *isc.AliasOutputWithID,
	incrementFactorOpt ...uint64,
) ([]state.Block, []*isc.AliasOutputWithID, []state.VirtualStateAccess) {
	var incrementFactor uint64
	if len(incrementFactorOpt) > 0 {
		incrementFactor = incrementFactorOpt[0]
	} else {
		incrementFactor = 1
	}
	result := make([]state.Block, count+1)
	vStates := make([]state.VirtualStateAccess, len(result))
	vStates[0] = vs
	aliasOutputs := make([]*isc.AliasOutputWithID, len(result))
	aliasOutputs[0] = aliasOutput
	for i := 1; i < len(result); i++ {
		baseIndex := (i + branchingFactor - 2) / branchingFactor
		increment := uint64(1+i%branchingFactor) * incrementFactor
		result[i], aliasOutputs[i], vStates[i] = GetNextState(t, vStates[baseIndex], aliasOutputs[baseIndex], increment)
	}
	return result[1:], aliasOutputs[1:], vStates[1:]
}

func GetNextState(
	t require.TestingT,
	vs state.VirtualStateAccess,
	consumedAliasOutputWithID *isc.AliasOutputWithID,
	incrementOpt ...uint64,
) (state.Block, *isc.AliasOutputWithID, state.VirtualStateAccess) {
	nextVS := vs.Copy()
	prevBlockIndex := vs.BlockIndex()
	counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")
	counterBin, err := nextVS.KVStore().Get(counterKey)
	require.NoError(t, err)
	counter, err := codec.DecodeUint64(counterBin, 0)
	require.NoError(t, err)
	var increment uint64
	if len(incrementOpt) > 0 {
		increment = incrementOpt[0]
	} else {
		increment = 1
	}

	consumedAliasOutput := consumedAliasOutputWithID.GetAliasOutput()
	vsCommitment, err := state.L1CommitmentFromBytes(consumedAliasOutput.StateMetadata)
	require.NoError(t, err)
	suBlockIndex := state.NewStateUpdateWithBlockLogValues(prevBlockIndex+1, time.Now(), &vsCommitment)

	suCounter := state.NewStateUpdate()
	counterBin = codec.EncodeUint64(counter + increment)
	suCounter.Mutations().Set(counterKey, counterBin)

	nextVS.ApplyStateUpdate(suBlockIndex)
	nextVS.ApplyStateUpdate(suCounter)
	nextVS.Commit()
	require.EqualValues(t, prevBlockIndex+1, nextVS.BlockIndex())

	block, err := nextVS.ExtractBlock()
	require.NoError(t, err)

	aliasOutput := &iotago.AliasOutput{
		Amount:         consumedAliasOutput.Amount,
		NativeTokens:   consumedAliasOutput.NativeTokens,
		AliasID:        consumedAliasOutput.AliasID,
		StateIndex:     consumedAliasOutput.StateIndex + 1,
		StateMetadata:  state.NewL1Commitment(trie.RootCommitment(nextVS.TrieNodeStore()), block.GetHash()).Bytes(),
		FoundryCounter: consumedAliasOutput.FoundryCounter,
		Conditions:     consumedAliasOutput.Conditions,
		Features:       consumedAliasOutput.Features,
	}
	aliasOutputID := iotago.OutputIDFromTransactionIDAndIndex(getRandomTxID(), 0).UTXOInput()
	aliasOutputWithID := isc.NewAliasOutputWithID(aliasOutput, aliasOutputID)

	return block, aliasOutputWithID, nextVS
}

func ContainsBlockHash(blockHash state.BlockHash, blockHashes []state.BlockHash) bool {
	for _, bh := range blockHashes {
		if bh.Equals(blockHash) {
			return true
		}
	}
	return false
}

func DeleteBlockHash(blockHash state.BlockHash, blockHashes []state.BlockHash) []state.BlockHash {
	for i := range blockHashes {
		if blockHashes[i].Equals(blockHash) {
			blockHashes[i] = blockHashes[len(blockHashes)-1]
			return blockHashes[:len(blockHashes)-1]
		}
	}
	return blockHashes
}

func RemoveAllBlockHashes(blockHashesToRemove []state.BlockHash, blockHashes []state.BlockHash) []state.BlockHash {
	result := blockHashes
	for i := range blockHashesToRemove {
		result = DeleteBlockHash(blockHashesToRemove[i], result)
	}
	return result
}

func AllDifferentBlockHashes(blockHashes []state.BlockHash) bool {
	for i := range blockHashes {
		for j := 0; j < i; j++ {
			if blockHashes[i].Equals(blockHashes[j]) {
				return false
			}
		}
	}
	return true
}

func getRandomTxID() iotago.TransactionID {
	var result iotago.TransactionID
	rand.Read(result[:])
	return result
}
