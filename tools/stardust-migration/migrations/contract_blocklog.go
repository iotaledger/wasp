package migrations

import (
	"log"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

func MigrateBlocklogContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft) {

	log.Print("Migrating blocklog contract\n")
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_blocklog.Contract.Hname())

	printWarningsForUnprocessableRequests(oldContractState)

	log.Print("Migrated blocklog contract\n")
}

func printWarningsForUnprocessableRequests(oldState old_kv.KVStoreReader) {
	// No need to migrate. Just print a warning if there are any

	log.Printf("Listing Unprocessable Requests...\n")

	count := IterateByPrefix(oldState, old_blocklog.PrefixUnprocessableRequests, func(k isc.RequestID, v []byte) {
		log.Printf("Warning: unprocessable request found %v", k.String())
	})

	log.Printf("Listing Unprocessable Requests completed (found %v records)\n", count)
}
