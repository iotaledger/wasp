package m002

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
)

var UpdateEVMISCMagic = migrations.Migration{
	Contract: evm.Contract,
	Apply: func(state kv.KVStore, log *logger.Logger) error {
		// noop - this migration was executed in the testnet, it had an issue where a bogus key was updated (not problematic)
		// keeping the code below if necessary to revisit
		return nil
		// evmPartition := subrealm.New(state, kv.Key(evm.Contract.Hname().Bytes())) // NOTE: this line was unnecessary
		// emulatorState := evm.EmulatorStateSubrealm(evmPartition)
		// stateDBSubrealm := emulator.StateDBSubrealm(emulatorState)
		// emulator.SetCode(stateDBSubrealm, iscmagic.ERC721NFTsAddress, iscmagic.ERC721NFTsRuntimeBytecode)
		// return nil
	},
}
