package validation

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
)

func OldStateContentToStr(chainState old_kv.KVStoreReader, chainID old_isc.ChainID, firstIndex, lastIndex uint32) string {
	var accountsContractStr, blocklogContractStr, evmContractStr string

	GoAllAndWait(func() {
		//accountsContractStr = oldAccountsContractContentToStr(chainState, chainID)
	}, func() {
		//blocklogContractStr = oldBlocklogContractContentToStr(chainState, firstIndex, lastIndex)
	}, func() {
		evmContractStr = oldEVMContractContentToStr(chainState, firstIndex, lastIndex)
	})

	// NOTE: for later states this mst likely will take huge amount of time and could cause OOM. For final testing need to change this flow.
	return accountsContractStr + blocklogContractStr + evmContractStr
}

func NewStateContentToStr(chainState kv.KVStoreReader, chainID isc.ChainID, firstIndex, lastIndex uint32) string {
	var accountsContractStr, blocklogContractStr, evmContractStr string
	GoAllAndWait(func() {
		//accountsContractStr = newAccountsContractContentToStr(chainState, chainID)
	}, func() {
		//blocklogContractStr = newBlocklogContractContentToStr(chainState, firstIndex, lastIndex)
	}, func() {
		evmContractStr = newEVMContractContentToStr(chainState, firstIndex, lastIndex)
	})

	// NOTE: for later states this mst likely will take huge amount of time and could cause OOM. For final testing need to change this flow.
	return accountsContractStr + blocklogContractStr + evmContractStr
}
