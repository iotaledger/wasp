package oldstate

import (
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_subrealm "github.com/nnikolash/wasp-types-exported/packages/kv/subrealm"
)

func GetContactStateReader(chainState old_kv.KVStoreReader, contractHname old_isc.Hname) old_kv.KVStoreReader {
	return old_subrealm.NewReadOnly(chainState, old_kv.Key(contractHname.Bytes()))
}
