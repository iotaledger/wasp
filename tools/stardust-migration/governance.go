package main

import (
	"log"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
)

func migrateGovernanceContract(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	panic("TODO: review")

	srcState := getContactStateReader(srcChainState, old_governance.Contract.Hname())
	destState := getContactState(destChainState, governance.Contract.Hname())

	log.Print("Migrating governance contract\n")

	// Chain Owner
	log.Printf("Migrating chain owner...\n")
	MigrateVariable(srcState, destState, old_governance.VarChainOwnerID, governance.VarChainOwnerID, OldAgentIDtoNewAgentID)
	log.Printf("Migrated chain owner\n")

	// Chain Owner delegated
	log.Printf("Migrating chain owner delegated...\n")
	MigrateVariable(srcState, destState, old_governance.VarChainOwnerIDDelegated, governance.VarChainOwnerIDDelegated, OldAgentIDtoNewAgentID)
	log.Printf("Migrated chain owner delegated\n")

	// Payout agent
	log.Printf("Migrating Payout agent...\n")
	MigrateVariable(srcState, destState, old_governance.VarPayoutAgentID, governance.VarPayoutAgentID, OldAgentIDtoNewAgentID)
	log.Printf("Migrated Payout agent\n")

	log.Print("Migrated governance contract\n")
}
