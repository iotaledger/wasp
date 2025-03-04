package newstate

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

func GetContactState(chainState kv.KVStore, contractHname isc.Hname) kv.KVStore {
	return subrealm.New(chainState, kv.Key(contractHname.Bytes()))
}

func GetContactStateReader(chainState kv.KVStoreReader, contractHname isc.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(contractHname.Bytes()))
}
