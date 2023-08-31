package accounts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func newFoundriesArray(state kv.KVStore) *collections.Array {
	return collections.NewArray(state, keyNewFoundries)
}

func accountFoundriesMap(state kv.KVStore, agentID isc.AgentID) *collections.Map {
	return collections.NewMap(state, foundriesMapKey(agentID))
}

func accountFoundriesMapR(state kv.KVStoreReader, agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, foundriesMapKey(agentID))
}

func AllFoundriesMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyFoundryOutputRecords)
}

func allFoundriesMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyFoundryOutputRecords)
}

// SaveFoundryOutput stores foundry output into the map of all foundry outputs (compressed form)
func SaveFoundryOutput(state kv.KVStore, f *iotago.FoundryOutput, outputIndex uint16) {
	foundryRec := foundryOutputRec{
		// TransactionID is unknown yet, will be filled next block
		OutputID:    iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, outputIndex),
		Amount:      f.Amount,
		TokenScheme: f.TokenScheme,
		Metadata:    []byte{},
	}
	AllFoundriesMap(state).SetAt(codec.EncodeUint32(f.SerialNumber), foundryRec.Bytes())
	newFoundriesArray(state).Push(codec.EncodeUint32(f.SerialNumber))
}

func updateFoundryOutputIDs(state kv.KVStore, anchorTxID iotago.TransactionID) {
	newFoundries := newFoundriesArray(state)
	allFoundries := AllFoundriesMap(state)
	n := newFoundries.Len()
	for i := uint32(0); i < n; i++ {
		k := newFoundries.GetAt(i)
		rec := mustFoundryOutputRecFromBytes(allFoundries.GetAt(k))
		rec.OutputID = iotago.OutputIDFromTransactionIDAndIndex(anchorTxID, rec.OutputID.Index())
		allFoundries.SetAt(k, rec.Bytes())
	}
	newFoundries.Erase()
}

// DeleteFoundryOutput deletes foundry output from the map of all foundries
func DeleteFoundryOutput(state kv.KVStore, sn uint32) {
	AllFoundriesMap(state).DelAt(codec.EncodeUint32(sn))
}

// GetFoundryOutput returns foundry output, its block number and output index
func GetFoundryOutput(state kv.KVStoreReader, sn uint32, chainID isc.ChainID) (*iotago.FoundryOutput, iotago.OutputID) {
	data := allFoundriesMapR(state).GetAt(codec.EncodeUint32(sn))
	if data == nil {
		return nil, iotago.OutputID{}
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
	return ret, rec.OutputID
}

// hasFoundry checks if specific account owns the foundry
func hasFoundry(state kv.KVStoreReader, agentID isc.AgentID, sn uint32) bool {
	return accountFoundriesMapR(state, agentID).HasAt(codec.EncodeUint32(sn))
}

// addFoundryToAccount ads new foundry to the foundries controlled by the account
func addFoundryToAccount(state kv.KVStore, agentID isc.AgentID, sn uint32) {
	key := codec.EncodeUint32(sn)
	foundries := accountFoundriesMap(state, agentID)
	if foundries.HasAt(key) {
		panic(ErrRepeatingFoundrySerialNumber)
	}
	foundries.SetAt(key, codec.EncodeBool(true))
}

func deleteFoundryFromAccount(state kv.KVStore, agentID isc.AgentID, sn uint32) {
	key := codec.EncodeUint32(sn)
	foundries := accountFoundriesMap(state, agentID)
	if !foundries.HasAt(key) {
		panic(ErrFoundryNotFound)
	}
	foundries.DelAt(key)
}

// MoveFoundryBetweenAccounts changes ownership of the foundry
func MoveFoundryBetweenAccounts(state kv.KVStore, agentIDFrom, agentIDTo isc.AgentID, sn uint32) {
	deleteFoundryFromAccount(state, agentIDFrom, sn)
	addFoundryToAccount(state, agentIDTo, sn)
}
