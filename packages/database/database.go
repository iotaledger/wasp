package database

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
)

func GetRegistryKVStore() kvstore.KVStore {
	return dbmanager.Instance().GetRegistryKVStore()
}

func GetOrCreateKVStore(chainID *ledgerstate.AliasAddress, dedicatedDbInstance bool) kvstore.KVStore {
	return dbmanager.Instance().GetOrCreateKVStore(chainID, dedicatedDbInstance)
}

func GetKVStore(chainID *ledgerstate.AliasAddress) kvstore.KVStore {
	return dbmanager.Instance().GetKVStore(chainID)
}
