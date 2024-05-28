// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog)

var (
	// Funcs
	FuncRetryUnprocessable = coreutil.Func("retryUnprocessable")

	// Views
	ViewGetBlockInfo               = coreutil.ViewFunc("getBlockInfo")
	ViewGetRequestIDsForBlock      = coreutil.ViewFunc("getRequestIDsForBlock")
	ViewGetRequestReceipt          = coreutil.ViewFunc("getRequestReceipt")
	ViewGetRequestReceiptsForBlock = coreutil.ViewFunc("getRequestReceiptsForBlock")
	ViewIsRequestProcessed         = coreutil.ViewFunc("isRequestProcessed")
	ViewGetEventsForRequest        = coreutil.ViewFunc("getEventsForRequest")
	ViewGetEventsForBlock          = coreutil.ViewFunc("getEventsForBlock")
	ViewHasUnprocessable           = coreutil.ViewFunc("hasUnprocessable")
)

// request parameters
const (
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

const (
	// Array of blockIndex => BlockInfo (pruned)
	// Covered in: TestGetEvents
	PrefixBlockRegistry = "a"

	// Map of request.ID().LookupDigest() => []RequestLookupKey (pruned)
	//   LookupDigest = reqID[:6] | outputIndex
	//   RequestLookupKey = blockIndex | requestIndex
	// Covered in: TestGetEvents
	prefixRequestLookupIndex = "b"

	// Map of RequestLookupKey => RequestReceipt (pruned)
	//   RequestLookupKey = blockIndex | requestIndex
	// Covered in: TestGetEvents
	prefixRequestReceipts = "c"

	// Map of EventLookupKey => event (pruned)
	//   EventLookupKey = blockIndex | requestIndex | eventIndex
	// Covered in: TestGetEvents
	prefixRequestEvents = "d"

	// Map of requestID => unprocessableRequestRecord
	// Covered in: TestUnprocessableWithPruning
	prefixUnprocessableRequests = "u"

	// Array of requestID.
	// Temporary list of unprocessable requests that need updating the outputID field
	// Covered in: TestUnprocessableWithPruning
	prefixNewUnprocessableRequests = "U"
)
