package runvm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/vm"
)

// runTheRequest is a wrapper which
// - check and validates the context
// - checks authorisations for protected requests
// - checks if requests are protected
// - handles request token
// - processes reward logic
// - redirects reserved request codes (is supported) to hardcoded processing
// - redirects not reserved codes (is supported) to SC VM
// - in case of something not correct:
//  --- sends all funds transferred by the request transaction to the SC address less reward back
//  --- the whole operation is NOP wrt the SC state.
func runTheRequest(ctx *vm.VMContext, processor vm.Processor) {
	// destroy token corresponding to request
	// NOTE: it is assumed here that balances contain all necessary request token balances
	// it is checked in the dispatcher.dispatchAddressUpdate
	err := ctx.TxBuilder.EraseColor(ctx.Address, (balance.Color)(ctx.Request.Tx.ID()), 1)
	if err != nil {
		// not enough balance for requests tokens
		// major inconsistency, it must be checked before
		ctx.Log.Panicf("something wrong with request token for reqid = %s. Not all requests were processed: %v",
			ctx.Request.RequestId().String(), err)
	}

}
