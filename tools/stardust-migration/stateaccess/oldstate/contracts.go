package oldstate

import (
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_subrealm "github.com/nnikolash/wasp-types-exported/packages/kv/subrealm"
)

func GetContactStateReader(chainState old_kv.KVStoreReader, contractHname old_isc.Hname) old_kv.KVStoreReader {
	return old_subrealm.NewReadOnly(chainState, old_kv.Key(contractHname.Bytes()))
}

func GetContactStateGetter(chainState old_kv.KVReader, contractHname old_isc.Hname) old_kv.KVReader {
	r := struct {
		old_kv.KVReader
		old_kv.KVIterator
	}{
		KVReader: chainState,
	}

	return old_subrealm.NewReadOnly(r, old_kv.Key(contractHname.Bytes()))
}

func GetContactStateIterator(chainState old_kv.KVIterator, contractHname old_isc.Hname) old_kv.KVIterator {
	r := struct {
		old_kv.KVReader
		old_kv.KVIterator
	}{
		KVIterator: chainState,
	}

	return old_subrealm.NewReadOnly(r, old_kv.Key(contractHname.Bytes()))
}
