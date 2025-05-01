package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/kvstore"

	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/dbinspector/stardustwasp"
)

func compareKey(contractKeys map[string]string, key kv.Key) string {
	for k, v := range contractKeys {
		keyLen := len(k)

		if keyLen > len(key) {
			continue
		}

		sliced := key[:keyLen]
		if bytes.Equal([]byte(k), []byte(sliced)) {
			return v
		}
	}

	return ""
}

func getContractStats(contract *coreutil.ContractInfo, keyMap map[string]string, state state.State) map[string]int {
	seenKeys := map[string]int{}

	// Check for Accounts
	state.IterateKeys(kv.Key(contract.Hname().Bytes()), func(key kv.Key) bool {
		// hn, _ := isc.HnameFromBytes([]byte(key[:4]))
		// name := corecontracts.All[hn]

		keyName := compareKey(keyMap, key[4:])
		if keyName == "" {
			// That shouldn't happen
			fmt.Printf("ERROR READING KEY %x\n", key)
			errKey := fmt.Sprintf("ERROR %s/%x", key, key)
			if _, ok := seenKeys[errKey]; !ok {
				seenKeys[errKey] = 0
			}
			seenKeys["ERROR"]++
		} else {
			if _, ok := seenKeys[keyName]; !ok {
				seenKeys[keyName] = 0
			}

			seenKeys[keyName]++
			// fmt.Printf("KEY:	%s, contract: %s, keyName: %s ->%x\n", accounts.Contract.Hname().String(), name.Name, keyName, key[4:])
		}

		return true
	})

	return seenKeys
}

func stateStatsPerHnameWithKeys(ctx context.Context, kvs kvstore.KVStore) {
	state := getState(kvs, blockIndex)

	fmt.Println("Accounts")
	for k, v := range getContractStats(accounts.Contract, stardustwasp.AccountsKeys, state) {
		fmt.Printf("\t%s: %d\n", k, v)
	}

	fmt.Println("Governance")
	for k, v := range getContractStats(governance.Contract, stardustwasp.GovernanceKeys, state) {
		fmt.Printf("\t%s: %d\n", k, v)
	}

	fmt.Println("EVM")
	for k, v := range getContractStats(evm.Contract, stardustwasp.EvmKeys, state) {
		fmt.Printf("\t%s: %d\n", k, v)
	}

	fmt.Println("BlockLog")
	for k, v := range getContractStats(blocklog.Contract, stardustwasp.BlocklogKeys, state) {
		fmt.Printf("\t%s: %d\n", k, v)
	}
}
