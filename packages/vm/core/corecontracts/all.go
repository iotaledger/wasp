package corecontracts

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

var All = map[iscp.Hname]*coreutil.ContractInfo{
	root.Contract.Hname():       root.Contract,
	errors.Contract.Hname():     errors.Contract,
	accounts.Contract.Hname():   accounts.Contract,
	blob.Contract.Hname():       blob.Contract,
	blocklog.Contract.Hname():   blocklog.Contract,
	governance.Contract.Hname(): governance.Contract,
	evm.Contract.Hname():        evm.Contract,
}

func IsCoreHname(hname iscp.Hname) bool {
	_, ret := All[hname]
	return ret
}
