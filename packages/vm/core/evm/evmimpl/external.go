package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
)

func Nonce(evmPartition kv.KVStoreReader, addr common.Address) uint64 {
	emuState := evm.EmulatorStateSubrealmR(evmPartition)
	stateDBStore := emulator.StateDBSubrealmR(emuState)
	return emulator.GetNonce(stateDBStore, addr)
}
