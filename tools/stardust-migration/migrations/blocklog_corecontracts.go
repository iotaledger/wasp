package migrations

import (
	"fmt"

	"github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"

	new_isc "github.com/iotaledger/wasp/packages/isc"
)

func migrateContractCall(contract isc.Hname, entrypoint isc.Hname, args dict.Dict) new_isc.CallArguments {
	if contract == governance.Contract.Hname() && entrypoint == governance.FuncSetFeePolicy.Hname() {
		return new_isc.NewCallArguments()
	}

	if contract == evm.Contract.Hname() && entrypoint == evm.FuncSendTransaction.Hname() {
		return new_isc.NewCallArguments()
	}

	if contract == accounts.Contract.Hname() && entrypoint == accounts.FuncTransferAllowanceTo.Hname() {
		return new_isc.NewCallArguments()
	}

	panic(fmt.Sprintf("failed to migrate contract call. Contract: %s, Entrypoint: %s unsupported!", contract, entrypoint))
}
