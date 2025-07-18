package vmimpl

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
)

const MaxPostedOutputsInOneRequest = 4

func (reqctx *requestContext) send(params isc.RequestParameters) {
	// simply send assets to a L1 address
	reqctx.vm.txbuilder.SendAssets(params.TargetAddress.AsIotaAddress(), params.Assets)

	account := reqctx.CurrentContractAccountID()
	reqctx.accountsStateWriter(false).DebitFromAccount(account, params.Assets.Coins)
	for obj := range params.Assets.Objects.Iterate() {
		reqctx.accountsStateWriter(false).DebitObjectFromAccount(account, obj.ID)
	}
}
