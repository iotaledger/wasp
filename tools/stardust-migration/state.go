package main

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_subrealm "github.com/nnikolash/wasp-types-exported/packages/kv/subrealm"
)

func getContactStateReader(chainState old_kv.KVStoreReader, contractHname old_isc.Hname) old_kv.KVStoreReader {
	return old_subrealm.NewReadOnly(chainState, old_kv.Key(contractHname.Bytes()))
}

func getContactState(chainState kv.KVStore, contractHname isc.Hname) kv.KVStore {
	return subrealm.New(chainState, kv.Key(contractHname.Bytes()))
}
