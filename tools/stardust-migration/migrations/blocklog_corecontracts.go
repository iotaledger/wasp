package migrations

import (
	"fmt"

	"github.com/samber/lo"

	old_evm_types "github.com/nnikolash/wasp-types-exported/packages/evm/evmtypes"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
	old_gas "github.com/nnikolash/wasp-types-exported/packages/vm/gas"

	new_isc "github.com/iotaledger/wasp/packages/isc"
	new_util "github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	new_evm "github.com/iotaledger/wasp/packages/vm/core/evm"
	new_governance "github.com/iotaledger/wasp/packages/vm/core/governance"
	new_gas "github.com/iotaledger/wasp/packages/vm/gas"
)

func migrateGovernanceSetFeePolicy(args old_kv.Dict) new_isc.Message {
	oldFeePolicy := lo.Must(old_gas.FeePolicyFromBytes(args.Get(old_governance.ParamFeePolicyBytes)))

	newGasPerToken := OldGasPerTokenToNew(oldFeePolicy)

	newFeePolicy := new_governance.FuncSetFeePolicy.Message(&new_gas.FeePolicy{
		EVMGasRatio: new_util.Ratio32{
			A: oldFeePolicy.EVMGasRatio.A,
			B: oldFeePolicy.EVMGasRatio.B,
		},
		GasPerToken:       newGasPerToken,
		ValidatorFeeShare: oldFeePolicy.ValidatorFeeShare,
	})

	return newFeePolicy
}

func migrateEVMSendTransaction(args old_kv.Dict) new_isc.Message {
	oldTransaction := lo.Must(old_evm_types.DecodeTransaction(args.Get(old_evm.FieldTransaction)))
	return new_evm.FuncSendTransaction.Message(oldTransaction)
}

func migrateAccountsTransferAllowanceTo(oldChainID old_isc.ChainID, args old_kv.Dict) new_isc.Message {
	agentID := lo.Must(old_isc.AgentIDFromBytes(args.Get(old_accounts.ParamAgentID)))
	return accounts.FuncTransferAllowanceTo.Message(OldAgentIDtoNewAgentID(agentID, oldChainID))
}

func migrateContractCall(oldChainID old_isc.ChainID, contract old_isc.Hname, entrypoint old_isc.Hname, args old_kv.Dict) new_isc.Message {
	if contract == old_governance.Contract.Hname() && entrypoint == old_governance.FuncSetFeePolicy.Hname() {
		return migrateGovernanceSetFeePolicy(args)
	}

	if contract == old_evm.Contract.Hname() && entrypoint == old_evm.FuncSendTransaction.Hname() {
		return migrateEVMSendTransaction(args)
	}

	if contract == old_accounts.Contract.Hname() && entrypoint == old_accounts.FuncTransferAllowanceTo.Hname() {
		return migrateAccountsTransferAllowanceTo(oldChainID, args)
	}

	panic(fmt.Sprintf("failed to migrate contract call. Contract: %s, Entrypoint: %s unsupported!", contract, entrypoint))
}
