package main

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
)

type migrationHandler func(key kv.Key, value []byte, destStore state.StateDraft) error

var prefixMigrationHandlers = map[string]migrationHandler{
	keyAllAccounts: accountsMigrationHandler,
}

func getMigrationHandler(prefix string) migrationHandler {
	handler, ok := prefixMigrationHandlers[prefix]
	if !ok {
		panic("no migration handler for prefix: " + prefix)
	}

	return handler
}

func accountsMigrationHandler(key kv.Key, value []byte, destStore state.StateDraft) error {
	// TODO: this is a temp value for POC
	zeroBytes := make([]byte, 8)
	destStore.Set(key, zeroBytes)
	return nil
}
