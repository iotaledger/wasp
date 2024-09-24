package isc_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func TestVMErrorCodeSerialization(t *testing.T) {
	bcs.TestCodecVsRef(t, isc.VMErrorCode{
		ContractID: isc.Hname(1074),
		ID:         123,
	})

	bcs.TestCodec(t, isc.UnresolvedVMError{
		ErrorCode: blocklog.ErrBlockNotFound.Code(),
		Params:    []isc.VMParam{int32(1), int8(2), "string"},
	})
}

func TestVMParamCodec(t *testing.T) {
	bcs.TestCodec(t, isc.VMParam(int8(123)))
	bcs.TestCodec(t, isc.VMParam(int16(123)))
	bcs.TestCodec(t, isc.VMParam(int32(123)))
	bcs.TestCodec(t, isc.VMParam(int64(123)))
	bcs.TestCodec(t, isc.VMParam(uint8(123)))
	bcs.TestCodec(t, isc.VMParam(uint16(123)))
	bcs.TestCodec(t, isc.VMParam(uint32(123)))
	bcs.TestCodec(t, isc.VMParam(uint64(123)))
	bcs.TestCodec(t, isc.VMParam("string"))
}
