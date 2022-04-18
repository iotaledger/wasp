// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
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
	FuncControlAddresses           = coreutil.ViewFunc("controlAddresses")
	FuncGetBlockInfo               = coreutil.ViewFunc("getBlockInfo")
	FuncGetLatestBlockInfo         = coreutil.ViewFunc("getLatestBlockInfo")
	FuncGetRequestIDsForBlock      = coreutil.ViewFunc("getRequestIDsForBlock")
	FuncGetRequestReceipt          = coreutil.ViewFunc("getRequestReceipt")
	FuncGetRequestReceiptsForBlock = coreutil.ViewFunc("getRequestReceiptsForBlock")
	FuncIsRequestProcessed         = coreutil.ViewFunc("isRequestProcessed")
	FuncGetEventsForRequest        = coreutil.ViewFunc("getEventsForRequest")
	FuncGetEventsForBlock          = coreutil.ViewFunc("getEventsForBlock")
	FuncGetEventsForContract       = coreutil.ViewFunc("getEventsForContract")
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
