package sctransaction

import "github.com/iotaledger/wasp/packages/util"

type RequestCode uint16

// user-defined request codes:
// unprotected from 0 to 2^14-1
// protected from 2^16 - 2^14 - 1
const (
	requestCodeReserved     = uint16(0x80)
	requestCodeProtected    = uint16(0x40)
	FirstBuiltInRequestCode = RequestCode(requestCodeReserved | requestCodeProtected)
)

func (rc RequestCode) IsProtected() bool {
	return uint16(rc)&(requestCodeProtected|requestCodeReserved) != 0
}

func (rc RequestCode) IsUserDefined() bool {
	return uint16(rc)&requestCodeReserved == 0
}

func (rc RequestCode) IsReserved() bool {
	return !rc.IsUserDefined()
}

func (rc RequestCode) Bytes() []byte {
	return util.Uint16To2Bytes(uint16(rc))
}
