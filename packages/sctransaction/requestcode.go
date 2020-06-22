package sctransaction

import "github.com/iotaledger/wasp/packages/util"

type RequestCode uint16

// user-defined request codes:
// unprotected from 0 to 2^14-1
// protected from 2^16 - 2^14 - 1
const (
	RequestCodeReserved          = uint16(0x8000)
	RequestCodeProtected         = uint16(0x4000)
	RequestCodeProtectedReserved = RequestCodeReserved | RequestCodeProtected
)

func (rc RequestCode) IsProtected() bool {
	return uint16(rc)&(RequestCodeProtected|RequestCodeReserved) != 0
}

func (rc RequestCode) IsUserDefined() bool {
	return uint16(rc)&RequestCodeReserved == 0
}

func (rc RequestCode) IsReserved() bool {
	return !rc.IsUserDefined()
}

func (rc RequestCode) Bytes() []byte {
	return util.Uint16To2Bytes(uint16(rc))
}
