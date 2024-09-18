package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
)

func eventRotate(ctx isc.Sandbox, newAddr *cryptolib.Address, oldAddr *cryptolib.Address) {
	evt := lo.T2(oldAddr, newAddr)
	ctx.Event("coregovernance.rotate", bcs.MustMarshal(&evt))
}
