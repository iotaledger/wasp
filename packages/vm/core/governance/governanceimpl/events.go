package governanceimpl

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func eventRotate(ctx isc.Sandbox, newAddr *cryptolib.Address, oldAddr *cryptolib.Address) {
	evt := lo.T2(oldAddr, newAddr)
	ctx.Event("coregovernance.rotate", bcs.MustMarshal(&evt))
}
