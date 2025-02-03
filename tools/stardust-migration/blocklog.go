package main

import (
	"log"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
)

func migrateBlocklogContract(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	panic("TODO: review")

	log.Print("Migrating blocklog contract\n")

	// No need to migrate. Just print a warning if there are any
	log.Printf("Listing Unprocessable Requests...\n")

	blocklogSrcState := getContactStateReader(srcChainState, old_blocklog.Contract.Hname())

	count := IterateByPrefix(blocklogSrcState, old_blocklog.PrefixUnprocessableRequests, func(k isc.RequestID, v []byte) {
		log.Printf("Warning: unprocessable request found %v", k.String())
	})

	log.Printf("Listing Unprocessable Requests completed (found %v records)\n", count)

	log.Print("Migrated blocklog contract\n")
}
