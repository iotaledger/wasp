// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog)

const (
	PrefixBlockRegistry      = "a"
	prefixRequestLookupIndex = "b"
	prefixRequestReceipts    = "c"
	prefixRequestEvents      = "d"

	// map of == request ID => unprocessableRequestRecord
	prefixUnprocessableRequests = "u"
	// array of request ID: list of unprocessable requests that
	// need updating the outputID field
	prefixNewUnprocessableRequests = "U"
)

var (
	ViewGetBlockInfo               = coreutil.ViewFunc("getBlockInfo")
	ViewGetRequestIDsForBlock      = coreutil.ViewFunc("getRequestIDsForBlock")
	ViewGetRequestReceipt          = coreutil.ViewFunc("getRequestReceipt")
	ViewGetRequestReceiptsForBlock = coreutil.ViewFunc("getRequestReceiptsForBlock")
	ViewIsRequestProcessed         = coreutil.ViewFunc("isRequestProcessed")
	ViewGetEventsForRequest        = coreutil.ViewFunc("getEventsForRequest")
	ViewGetEventsForBlock          = coreutil.ViewFunc("getEventsForBlock")
	ViewGetEventsForContract       = coreutil.ViewFunc("getEventsForContract")
	ViewHasUnprocessable           = coreutil.ViewFunc("hasUnprocessable")

	// entrypoints
	FuncRetryUnprocessable = coreutil.Func("retryUnprocessable")
)

const (
	// parameters
	ParamBlockIndex                 = "n"
	ParamBlockInfo                  = "i"
	ParamContractHname              = "h"
	ParamFromBlock                  = "f"
	ParamToBlock                    = "t"
	ParamRequestID                  = "u"
	ParamRequestIndex               = "r"
	ParamRequestProcessed           = "p"
	ParamRequestRecord              = "d"
	ParamEvent                      = "e"
	ParamStateControllerAddress     = "s"
	ParamUnprocessableRequestExists = "x"
)
