package main

import (
	"log"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"
)

func migrateBlocklogContract(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	panic("TODO: review")

	log.Print("Migrating blocklog contract\n")

	// Unprocessable Requests (blocklog contract)
	// No need to migrate. Just print a warning if there are any
	log.Printf("Listing Unprocessable Requests...\n")

	blocklogContractStateSrc := getContactStateReader(srcChainState, old_blocklog.Contract.Hname())
	count := 0
	old_collections.NewMapReadOnly(blocklogContractStateSrc, old_blocklog.PrefixUnprocessableRequests).Iterate(func(srcKey, srcBytes []byte) bool {
		reqID := lo.Must(Deserialize[isc.RequestID](srcKey))
		log.Printf("Warning: unprocessable request found %v", reqID.String())
		count++
		return true
	})

	log.Printf("Listing Unprocessable Requests completed (found %v records)\n", count)

	log.Print("Migrated blocklog contract\n")
}
