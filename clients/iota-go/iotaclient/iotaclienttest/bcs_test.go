package iotaclienttest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestUnmarshalBCS(t *testing.T) {
	v := "hello"
	vEnc := bcs.MustMarshal(&v)

	var vDec string
	err := iotaclient.UnmarshalBCS(vEnc, &vDec)
	require.NoError(t, err)
	require.Equal(t, v, vDec)

	vDec = ""
	vEncWithExcess := append(vEnc, 0x1)
	err = iotaclient.UnmarshalBCS(vEncWithExcess, &vDec)
	require.Error(t, err)
}
