// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog, "Block log contract")

const (
	prefixBlockRegistry = string('a' + iota)
	prefixControlAddresses
	prefixRequestLookupIndex
	prefixRequestReceipts
	prefixRequestEvents
	prefixSmartContractEventsLookup
)

var (
	ViewControlAddresses           = coreutil.ViewFunc("controlAddresses")
	ViewGetBlockInfo               = coreutil.ViewFunc("getBlockInfo")
	ViewGetRequestIDsForBlock      = coreutil.ViewFunc("getRequestIDsForBlock")
	ViewGetRequestReceipt          = coreutil.ViewFunc("getRequestReceipt")
	ViewGetRequestReceiptsForBlock = coreutil.ViewFunc("getRequestReceiptsForBlock")
	ViewIsRequestProcessed         = coreutil.ViewFunc("isRequestProcessed")
	ViewGetEventsForRequest        = coreutil.ViewFunc("getEventsForRequest")
	ViewGetEventsForBlock          = coreutil.ViewFunc("getEventsForBlock")
	ViewGetEventsForContract       = coreutil.ViewFunc("getEventsForContract")
)

const (
	// parameters
	ParamBlockIndex             = "n"
	ParamBlockInfo              = "i"
	ParamGoverningAddress       = "g"
	ParamContractHname          = "h"
	ParamFromBlock              = "f"
	ParamToBlock                = "t"
	ParamRequestID              = "u"
	ParamRequestIndex           = "r"
	ParamRequestProcessed       = "p"
	ParamRequestRecord          = "d"
	ParamEvent                  = "e"
	ParamStateControllerAddress = "s"
)
