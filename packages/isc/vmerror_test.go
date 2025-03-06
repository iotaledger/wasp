package isc_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func TestVMErrorCodeSerialization(t *testing.T) {
	bcs.TestCodec(t, isc.VMErrorCode{
		ContractID: isc.Hname(1074),
		ID:         123,
	})

	bcs.TestCodec(t, isc.UnresolvedVMError{
		ErrorCode: blocklog.ErrBlockNotFound.Code(),
		Params:    []isc.VMErrorParam{int32(1), int8(2), "string"},
	})
}

func TestVMParamCodec(t *testing.T) {
	bcs.TestCodec(t, isc.VMErrorParam(int8(123)))
	bcs.TestCodec(t, isc.VMErrorParam(int16(123)))
	bcs.TestCodec(t, isc.VMErrorParam(int32(123)))
	bcs.TestCodec(t, isc.VMErrorParam(int64(123)))
	bcs.TestCodec(t, isc.VMErrorParam(uint8(123)))
	bcs.TestCodec(t, isc.VMErrorParam(uint16(123)))
	bcs.TestCodec(t, isc.VMErrorParam(uint32(123)))
	bcs.TestCodec(t, isc.VMErrorParam(uint64(123)))
	bcs.TestCodec(t, isc.VMErrorParam("string"))
}
