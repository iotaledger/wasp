package accounts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func accountFoundriesMap(state kv.KVStore, agentID isc.AgentID) *collections.Map {
	return collections.NewMap(state, foundriesMapKey(agentID))
}

func accountFoundriesMapR(state kv.KVStoreReader, agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, foundriesMapKey(agentID))
}

func allFoundriesMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyFoundryOutputRecords)
}

func allFoundriesMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyFoundryOutputRecords)
}

// SaveFoundryOutput stores foundry output into the map of all foundry outputs (compressed form)
func SaveFoundryOutput(state kv.KVStore, f *iotago.FoundryOutput, blockIndex uint32, outputIndex uint16) {
	foundryRec := foundryOutputRec{
		Amount:      f.Amount,
		TokenScheme: f.TokenScheme,
		Metadata:    []byte{},
		BlockIndex:  blockIndex,
		OutputIndex: outputIndex,
	}
	allFoundriesMap(state).MustSetAt(codec.EncodeUint32(f.SerialNumber), foundryRec.Bytes())
}

// DeleteFoundryOutput deletes foundry output from the map of all foundries
func DeleteFoundryOutput(state kv.KVStore, sn uint32) {
	allFoundriesMap(state).MustDelAt(codec.EncodeUint32(sn))
}

// GetFoundryOutput returns foundry output, its block number and output index
func GetFoundryOutput(state kv.KVStoreReader, sn uint32, chainID isc.ChainID) (*iotago.FoundryOutput, uint32, uint16) {
	data := allFoundriesMapR(state).MustGetAt(codec.EncodeUint32(sn))
	if data == nil {
		return nil, 0, 0
	}
	rec := mustFoundryOutputRecFromBytes(data)

	ret := &iotago.FoundryOutput{
		Amount:       rec.Amount,
		NativeTokens: nil,
		SerialNumber: sn,
		TokenScheme:  rec.TokenScheme,
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: chainID.AsAddress().(*iotago.AliasAddress)},
		},
		Features: nil,
	}
	return ret, rec.BlockIndex, rec.OutputIndex
}

// hasFoundry checks if specific account owns the foundry
func hasFoundry(state kv.KVStoreReader, agentID isc.AgentID, sn uint32) bool {
	return accountFoundriesMapR(state, agentID).MustHasAt(codec.EncodeUint32(sn))
}

// addFoundryToAccount ads new foundry to the foundries controlled by the account
func addFoundryToAccount(state kv.KVStore, agentID isc.AgentID, sn uint32) {
	key := codec.EncodeUint32(sn)
	foundries := accountFoundriesMap(state, agentID)
	if foundries.MustHasAt(key) {
		panic(ErrRepeatingFoundrySerialNumber)
	}
	foundries.MustSetAt(key, codec.EncodeBool(true))
}

func deleteFoundryFromAccount(state kv.KVStore, agentID isc.AgentID, sn uint32) {
	key := codec.EncodeUint32(sn)
	foundries := accountFoundriesMap(state, agentID)
	if !foundries.MustHasAt(key) {
		panic(ErrFoundryNotFound)
	}
	foundries.MustDelAt(key)
}

// MoveFoundryBetweenAccounts changes ownership of the foundry
func MoveFoundryBetweenAccounts(state kv.KVStore, agentIDFrom, agentIDTo isc.AgentID, sn uint32) {
	deleteFoundryFromAccount(state, agentIDFrom, sn)
	addFoundryToAccount(state, agentIDTo, sn)
}
