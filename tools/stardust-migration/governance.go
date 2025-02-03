package main

import (
	"log"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
	"github.com/samber/lo"
)

func migrateGovernanceContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft, newChainID isc.ChainID) {
	panic("TODO: implement")

	oldContractState := getContactStateReader(oldChainState, old_governance.Contract.Hname())
	newContractState := getContactState(newChainState, governance.Contract.Hname())

	log.Print("Migrating governance contract\n")

	migrateChainOwnerID(oldChainState, newContractState, newChainID) // WARNING: oldChainState is specifically used here
	migrateChainOwnerIDDelegetaed(oldContractState, newContractState, newChainID)
	migratePayoutAgent(oldContractState, newContractState, newChainID)

	log.Print("Migrated governance contract\n")
}

func migrateChainOwnerID(oldChainState old_kv.KVStoreReader, newContractState kv.KVStore, newChainID isc.ChainID) {
	log.Print("Migrating chain owner...\n")

	oldChainOwnerID := old_governance.NewStateAccess(oldChainState).ChainOwnerID()
	newChainOwnerID := OldAgentIDtoNewAgentID(oldChainOwnerID, newChainID)
	governance.NewStateWriter(newContractState).SetChainOwnerID(newChainOwnerID)

	log.Print("Migrated chain owner\n")
}

func migrateChainOwnerIDDelegetaed(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore, newChainID isc.ChainID) {
	log.Print("Migrating chain owner delegated...\n")

	oldChainOwnerDelegatedIDBytes := oldContractState.Get(old_governance.VarChainOwnerIDDelegated)
	if len(oldChainOwnerDelegatedIDBytes) != 0 {
		oldChainOwnerDelegatedID := lo.Must(old_codec.DecodeAgentID(oldChainOwnerDelegatedIDBytes))
		newChainIDOwnerDelegatedID := OldAgentIDtoNewAgentID(oldChainOwnerDelegatedID, newChainID)
		governance.NewStateWriter(newContractState).SetChainOwnerIDDelegated(newChainIDOwnerDelegatedID)
	}

	log.Print("Migrated chain owner delegated\n")
}

func migratePayoutAgent(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore, newChainID isc.ChainID) {
	log.Printf("Migrating Payout agent...\n")

	oldPayoudAgentID := old_governance.MustGetPayoutAgentID(oldContractState)
	newPayoutAgentID := OldAgentIDtoNewAgentID(oldPayoudAgentID, newChainID)

	governance.NewStateWriter(newContractState).SetPayoutAgentID(newPayoutAgentID)

	log.Printf("Migrated Payout agent\n")
}
