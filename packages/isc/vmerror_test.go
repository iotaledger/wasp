package isc_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func TestVMErrorCodeSerialization(t *testing.T) {
	bcs.TestCodecAndHash(t, isc.VMErrorCode{
		ContractID: isc.Hname(1074),
		ID:         123,
	}, "e4239733f6c1")

	bcs.TestCodecAndHash(t, isc.UnresolvedVMError{
		ErrorCode: blocklog.ErrBlockNotFound.Code(),
		Params:    []isc.VMErrorParam{int32(1), int8(2), "string"},
	}, "c90d63542477")
}

func TestVMParamCodec(t *testing.T) {
	bcs.TestCodecAndHash(t, isc.VMErrorParam(int8(123)), "c0dbc40f96f9")
	bcs.TestCodecAndHash(t, isc.VMErrorParam(int16(123)), "4e2ca5620b49")
	bcs.TestCodecAndHash(t, isc.VMErrorParam(int32(123)), "73dad35f83a0")
	bcs.TestCodecAndHash(t, isc.VMErrorParam(int64(123)), "08bc8c7ec2d0")
	bcs.TestCodecAndHash(t, isc.VMErrorParam(uint8(123)), "aaff75313459")
	bcs.TestCodecAndHash(t, isc.VMErrorParam(uint16(123)), "da2aee695769")
	bcs.TestCodecAndHash(t, isc.VMErrorParam(uint32(123)), "fd3dec34f885")
	bcs.TestCodecAndHash(t, isc.VMErrorParam(uint64(123)), "cb91b924b8b5")
	bcs.TestCodecAndHash(t, isc.VMErrorParam("string"), "c8520f62d44e")
}
