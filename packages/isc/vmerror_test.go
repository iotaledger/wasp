package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
)

func TestVMErrorCodeSerialization(t *testing.T) {
	vmerrTest := isc.VMErrorCode{
		ContractID: isc.Hname(1074),
		ID:         123,
	}
	data := vmerrTest.Bytes()
	vmerr, err := isc.VMErrorCodeFromBytes(data)
	require.NoError(t, err)
	require.Equal(t, vmerrTest, vmerr)
}
