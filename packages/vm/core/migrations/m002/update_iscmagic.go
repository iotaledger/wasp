package m002

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
)

var UpdateEVMISCMagic = migrations.Migration{
	Contract: evm.Contract,
	Apply: func(state kv.KVStore, log *logger.Logger) error {
		evmPartition := subrealm.New(state, kv.Key(evm.Contract.Hname().Bytes()))
		emulatorState := evm.EmulatorStateSubrealm(evmPartition)
		stateDBSubrealm := emulator.StateDBSubrealm(emulatorState)
		emulator.SetCode(stateDBSubrealm, iscmagic.ERC721NFTsAddress, iscmagic.ERC721NFTsRuntimeBytecode)
		return nil
	},
}
