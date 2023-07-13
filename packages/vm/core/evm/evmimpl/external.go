package evmimpl

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
)

func Nonce(evmPartition kv.KVStoreReader, addr common.Address) uint64 {
	evmState := evm.EmulatorStateSubrealmR(evmPartition)
	stateDBStore := emulator.StateDBSubrealmR(evmState)
	return emulator.GetNonce(stateDBStore, addr)
}

func CheckNonce(evmPartition kv.KVStore, addr common.Address, nonce uint64) error {
	expected := Nonce(evmPartition, addr)
	if nonce != expected {
		return fmt.Errorf("Invalid nonce, expected %d, got %d", expected, nonce)
	}
	return nil
}
