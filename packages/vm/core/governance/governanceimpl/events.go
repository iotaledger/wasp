package governanceimpl

import (
	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func eventRotate(ctx isc.Sandbox, newAddr *cryptolib.Address, oldAddr *cryptolib.Address) {
	evt := lo.T2(oldAddr, newAddr)
	ctx.Event("coregovernance.rotate", bcs.MustMarshal(&evt))
}
