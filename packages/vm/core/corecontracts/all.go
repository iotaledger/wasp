package corecontracts

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
)

var All = map[isc.Hname]*coreutil.ContractInfo{
	root.Contract.Hname():       root.Contract,
	errors.Contract.Hname():     errors.Contract,
	accounts.Contract.Hname():   accounts.Contract,
	blocklog.Contract.Hname():   blocklog.Contract,
	governance.Contract.Hname(): governance.Contract,
	evm.Contract.Hname():        evm.Contract,
}

func IsCoreHname(hname isc.Hname) bool {
	_, ok := All[hname]
	return ok
}
